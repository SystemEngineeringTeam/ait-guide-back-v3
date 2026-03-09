package postgres

import (
	"context"
	"fmt"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/model"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RoomRepository はPostgreSQLでの部屋リポジトリ実装。
type RoomRepository struct {
	pool *pgxpool.Pool
}

// NewRoomRepository は新しいRoomRepositoryを作成する。
func NewRoomRepository(pool *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{pool: pool}
}

func (r *RoomRepository) FindAll(ctx context.Context, filter repository.RoomFilter) ([]model.Room, error) {
	query := `SELECT id, room_id, building_id, node_id, name, description, floor, capacity, room_type FROM rooms WHERE 1=1`
	args := []any{}
	argIdx := 1

	if filter.BuildingID != nil {
		query += fmt.Sprintf(` AND building_id = $%d`, argIdx)
		args = append(args, *filter.BuildingID)
		argIdx++
	}
	if filter.Floor != nil {
		query += fmt.Sprintf(` AND floor = $%d`, argIdx)
		args = append(args, *filter.Floor)
		argIdx++
	}
	if filter.RoomType != nil {
		query += fmt.Sprintf(` AND room_type = $%d`, argIdx)
		args = append(args, *filter.RoomType)
		argIdx++
	}

	query += ` ORDER BY id`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query rooms: %w", err)
	}
	defer rows.Close()

	var rooms []model.Room
	for rows.Next() {
		var rm model.Room
		if err := rows.Scan(&rm.ID, &rm.RoomID, &rm.BuildingID, &rm.NodeID, &rm.Name, &rm.Description, &rm.Floor, &rm.Capacity, &rm.RoomType); err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, rm)
	}

	return rooms, nil
}

func (r *RoomRepository) FindByRoomID(ctx context.Context, roomID string) (*model.Room, error) {
	var rm model.Room
	err := r.pool.QueryRow(ctx,
		`SELECT id, room_id, building_id, node_id, name, description, floor, capacity, room_type FROM rooms WHERE room_id = $1`, roomID).
		Scan(&rm.ID, &rm.RoomID, &rm.BuildingID, &rm.NodeID, &rm.Name, &rm.Description, &rm.Floor, &rm.Capacity, &rm.RoomType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query room: %w", err)
	}
	return &rm, nil
}
