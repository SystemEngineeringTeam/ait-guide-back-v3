// Package loader はCSVファイルからのデータ読み込みを提供する。
package loader

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CSVSeeder はCSVファイルからDBにデータを投入する。
type CSVSeeder struct {
	pool *pgxpool.Pool
}

// NewCSVSeeder は新しいCSVSeederを作成する。
func NewCSVSeeder(pool *pgxpool.Pool) *CSVSeeder {
	return &CSVSeeder{pool: pool}
}

// SeedAll は指定ディレクトリ内のCSVファイルを順番に投入する。
func (s *CSVSeeder) SeedAll(ctx context.Context, seedDir string) error {
	order := []struct {
		file  string
		table string
		seed  func(ctx context.Context, path string) error
	}{
		{"buildings.csv", "buildings", s.SeedBuildings},
		{"nodes.csv", "nodes", s.SeedNodes},
		{"edges.csv", "edges", s.SeedEdges},
		{"rooms.csv", "rooms", s.SeedRooms},
		{"photos.csv", "photos", s.SeedPhotos},
	}

	for _, o := range order {
		path := filepath.Join(seedDir, o.file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("seed skipped (not found): %s\n", o.file)
			continue
		}
		if err := o.seed(ctx, path); err != nil {
			return fmt.Errorf("failed to seed %s: %w", o.file, err)
		}
		fmt.Printf("seeded: %s\n", o.file)
	}

	return nil
}

// SeedBuildings はbuildings.csvを投入する。
// CSV形式: id,name,description,affiliation
func (s *CSVSeeder) SeedBuildings(ctx context.Context, path string) error {
	records, err := readCSV(path)
	if err != nil {
		return err
	}

	for i, r := range records {
		if len(r) < 2 {
			return fmt.Errorf("line %d: insufficient columns", i+2)
		}
		id, err := strconv.Atoi(r[0])
		if err != nil {
			return fmt.Errorf("line %d: invalid id: %w", i+2, err)
		}
		name := r[1]
		description := nullableStr(r, 2)
		affiliation := nullableStr(r, 3)

		_, err = s.pool.Exec(ctx,
			`INSERT INTO buildings (id, name, description, affiliation) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING`,
			id, name, description, affiliation)
		if err != nil {
			return fmt.Errorf("line %d: %w", i+2, err)
		}
	}
	return nil
}

// SeedNodes はnodes.csvを投入する。
// CSV形式: node_id,lat,lng,node_type,building_id,floor
func (s *CSVSeeder) SeedNodes(ctx context.Context, path string) error {
	records, err := readCSV(path)
	if err != nil {
		return err
	}

	for i, r := range records {
		if len(r) < 4 {
			return fmt.Errorf("line %d: insufficient columns", i+2)
		}
		nodeID := r[0]
		lat, err := strconv.ParseFloat(r[1], 64)
		if err != nil {
			return fmt.Errorf("line %d: invalid lat: %w", i+2, err)
		}
		lng, err := strconv.ParseFloat(r[2], 64)
		if err != nil {
			return fmt.Errorf("line %d: invalid lng: %w", i+2, err)
		}
		nodeType := r[3]
		buildingID := nullableInt(r, 4)
		floor := nullableInt(r, 5)

		_, err = s.pool.Exec(ctx,
			`INSERT INTO nodes (node_id, geom, node_type, building_id, floor) VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), $4, $5, $6) ON CONFLICT (node_id) DO NOTHING`,
			nodeID, lng, lat, nodeType, buildingID, floor)
		if err != nil {
			return fmt.Errorf("line %d: %w", i+2, err)
		}
	}
	return nil
}

// SeedEdges はedges.csvを投入する。
// CSV形式: node_id_from,node_id_target,distance,cost,level,has_stairs,is_accessible,is_indoor
func (s *CSVSeeder) SeedEdges(ctx context.Context, path string) error {
	records, err := readCSV(path)
	if err != nil {
		return err
	}

	for i, r := range records {
		if len(r) < 8 {
			return fmt.Errorf("line %d: insufficient columns", i+2)
		}
		nodeFrom := r[0]
		nodeTarget := r[1]
		distance, err := strconv.ParseFloat(r[2], 64)
		if err != nil {
			return fmt.Errorf("line %d: invalid distance: %w", i+2, err)
		}
		cost := nullableFloat(r, 3)
		level, err := strconv.Atoi(r[4])
		if err != nil {
			return fmt.Errorf("line %d: invalid level: %w", i+2, err)
		}
		hasStairs := parseBool(r[5])
		isAccessible := parseBool(r[6])
		isIndoor := parseBool(r[7])

		_, err = s.pool.Exec(ctx,
			`INSERT INTO edges (node_id_from, node_id_target, distance, cost, level, has_stairs, is_accessible, is_indoor) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			nodeFrom, nodeTarget, distance, cost, level, hasStairs, isAccessible, isIndoor)
		if err != nil {
			return fmt.Errorf("line %d: %w", i+2, err)
		}
	}
	return nil
}

// SeedRooms はrooms.csvを投入する。
// CSV形式: room_id,building_id,node_id,name,description,floor,capacity,room_type
func (s *CSVSeeder) SeedRooms(ctx context.Context, path string) error {
	records, err := readCSV(path)
	if err != nil {
		return err
	}

	for i, r := range records {
		if len(r) < 2 {
			return fmt.Errorf("line %d: insufficient columns", i+2)
		}
		roomID := r[0]
		buildingID, err := strconv.Atoi(r[1])
		if err != nil {
			return fmt.Errorf("line %d: invalid building_id: %w", i+2, err)
		}
		nodeID := nullableStr(r, 2)
		name := nullableStr(r, 3)
		description := nullableStr(r, 4)
		floor := nullableInt(r, 5)
		capacity := nullableInt(r, 6)
		roomType := nullableStr(r, 7)

		_, err = s.pool.Exec(ctx,
			`INSERT INTO rooms (room_id, building_id, node_id, name, description, floor, capacity, room_type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (room_id) DO NOTHING`,
			roomID, buildingID, nodeID, name, description, floor, capacity, roomType)
		if err != nil {
			return fmt.Errorf("line %d: %w", i+2, err)
		}
	}
	return nil
}

// SeedPhotos はphotos.csvを投入する。
// CSV形式: building_id,url,caption,display_order
func (s *CSVSeeder) SeedPhotos(ctx context.Context, path string) error {
	records, err := readCSV(path)
	if err != nil {
		return err
	}

	for i, r := range records {
		if len(r) < 4 {
			return fmt.Errorf("line %d: insufficient columns", i+2)
		}
		buildingID, err := strconv.Atoi(r[0])
		if err != nil {
			return fmt.Errorf("line %d: invalid building_id: %w", i+2, err)
		}
		url := r[1]
		caption := nullableStr(r, 2)
		displayOrder, err := strconv.Atoi(r[3])
		if err != nil {
			return fmt.Errorf("line %d: invalid display_order: %w", i+2, err)
		}

		_, err = s.pool.Exec(ctx,
			`INSERT INTO photos (building_id, url, caption, display_order) VALUES ($1, $2, $3, $4)`,
			buildingID, url, caption, displayOrder)
		if err != nil {
			return fmt.Errorf("line %d: %w", i+2, err)
		}
	}
	return nil
}

func readCSV(path string) (_ [][]string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close %s: %w", path, cerr)
		}
	}()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	if len(records) < 2 {
		return nil, nil
	}
	return records[1:], nil // skip header
}

func nullableStr(r []string, idx int) *string {
	if idx >= len(r) || strings.TrimSpace(r[idx]) == "" {
		return nil
	}
	s := strings.TrimSpace(r[idx])
	return &s
}

func nullableInt(r []string, idx int) *int {
	if idx >= len(r) || strings.TrimSpace(r[idx]) == "" {
		return nil
	}
	v, err := strconv.Atoi(strings.TrimSpace(r[idx]))
	if err != nil {
		return nil
	}
	return &v
}

func nullableFloat(r []string, idx int) *float64 {
	if idx >= len(r) || strings.TrimSpace(r[idx]) == "" {
		return nil
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(r[idx]), 64)
	if err != nil {
		return nil
	}
	return &v
}

func parseBool(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "true" || s == "1" || s == "yes"
}
