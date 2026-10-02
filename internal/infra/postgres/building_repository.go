package postgres

import (
	"context"
	"fmt"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BuildingRepository はPostgreSQLでの建物リポジトリ実装。
type BuildingRepository struct {
	pool *pgxpool.Pool
}

// NewBuildingRepository は新しいBuildingRepositoryを作成する。
func NewBuildingRepository(pool *pgxpool.Pool) *BuildingRepository {
	return &BuildingRepository{pool: pool}
}

func (r *BuildingRepository) FindAll(ctx context.Context, limit int, affiliation *string) ([]model.Building, error) {
	query := `SELECT building_no, key, name, description, affiliation FROM buildings WHERE 1=1`
	args := []any{}
	argIdx := 1

	if affiliation != nil {
		query += fmt.Sprintf(` AND affiliation = $%d`, argIdx)
		args = append(args, *affiliation)
		argIdx++
	}

	query += fmt.Sprintf(` ORDER BY building_no LIMIT $%d`, argIdx)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query buildings: %w", err)
	}
	defer rows.Close()

	var buildings []model.Building
	for rows.Next() {
		var b model.Building
		if err := rows.Scan(&b.ID, &b.Key, &b.Name, &b.Description, &b.Affiliation); err != nil {
			return nil, fmt.Errorf("failed to scan building: %w", err)
		}
		buildings = append(buildings, b)
	}

	return buildings, nil
}

func (r *BuildingRepository) FindByID(ctx context.Context, id int) (*model.Building, error) {
	var b model.Building
	err := r.pool.QueryRow(ctx,
		`SELECT building_no, key, name, description, affiliation FROM buildings WHERE building_no = $1`, id).
		Scan(&b.ID, &b.Key, &b.Name, &b.Description, &b.Affiliation)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query building: %w", err)
	}

	// 写真を取得
	rows, err := r.pool.Query(ctx,
		`SELECT id, building_id, url, caption, display_order FROM photos WHERE building_id = $1 ORDER BY display_order`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query photos: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p model.Photo
		if err := rows.Scan(&p.ID, &p.BuildingID, &p.URL, &p.Caption, &p.DisplayOrder); err != nil {
			return nil, fmt.Errorf("failed to scan photo: %w", err)
		}
		b.Photos = append(b.Photos, p)
	}

	return &b, nil
}
