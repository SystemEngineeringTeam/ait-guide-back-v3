package postgres

import (
	"context"
	"fmt"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/model"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NodeRepository はPostgreSQLでのノードリポジトリ実装。
type NodeRepository struct {
	pool *pgxpool.Pool
}

// NewNodeRepository は新しいNodeRepositoryを作成する。
func NewNodeRepository(pool *pgxpool.Pool) *NodeRepository {
	return &NodeRepository{pool: pool}
}

func (r *NodeRepository) FindAll(ctx context.Context, filter repository.NodeFilter) ([]model.Node, error) {
	query := `SELECT id, node_id, ST_Y(geom) AS lat, ST_X(geom) AS lng, node_type, building_id, floor FROM nodes WHERE 1=1`
	args := []any{}
	argIdx := 1

	if filter.NodeID != nil {
		query += fmt.Sprintf(` AND node_id = $%d`, argIdx)
		args = append(args, *filter.NodeID)
		argIdx++
	}
	if filter.NodeType != nil {
		query += fmt.Sprintf(` AND node_type = $%d`, argIdx)
		args = append(args, string(*filter.NodeType))
		argIdx++
	}
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

	query += fmt.Sprintf(` ORDER BY id LIMIT $%d`, argIdx)
	args = append(args, filter.Limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query nodes: %w", err)
	}
	defer rows.Close()

	var nodes []model.Node
	for rows.Next() {
		var n model.Node
		if err := rows.Scan(&n.ID, &n.NodeID, &n.Lat, &n.Lng, &n.NodeType, &n.BuildingID, &n.Floor); err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}
		nodes = append(nodes, n)
	}

	return nodes, nil
}

func (r *NodeRepository) FindNearestByLatLng(ctx context.Context, lat, lng float64) (*model.Node, error) {
	var n model.Node
	err := r.pool.QueryRow(ctx,
		`SELECT id, node_id, ST_Y(geom) AS lat, ST_X(geom) AS lng, node_type, building_id, floor
		 FROM nodes
		 ORDER BY geom <-> ST_SetSRID(ST_MakePoint($1, $2), 4326)
		 LIMIT 1`,
		lng, lat).
		Scan(&n.ID, &n.NodeID, &n.Lat, &n.Lng, &n.NodeType, &n.BuildingID, &n.Floor)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find nearest node: %w", err)
	}
	return &n, nil
}

func (r *NodeRepository) FindNearestTargetByBuildingID(ctx context.Context, buildingID int, lat, lng float64) (*model.Node, error) {
	var n model.Node
	err := r.pool.QueryRow(ctx,
		`SELECT id, node_id, ST_Y(geom) AS lat, ST_X(geom) AS lng, node_type, building_id, floor
		 FROM nodes
		 WHERE node_type IN ('entrance', 'facility') AND building_id = $1
		 ORDER BY geom <-> ST_SetSRID(ST_MakePoint($2, $3), 4326)
		 LIMIT 1`,
		buildingID, lng, lat).
		Scan(&n.ID, &n.NodeID, &n.Lat, &n.Lng, &n.NodeType, &n.BuildingID, &n.Floor)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find nearest target node: %w", err)
	}
	return &n, nil
}
