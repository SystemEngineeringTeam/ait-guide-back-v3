package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RouteRepository はpgRoutingを使った経路探索リポジトリ実装。
type RouteRepository struct {
	pool *pgxpool.Pool
}

// NewRouteRepository は新しいRouteRepositoryを作成する。
func NewRouteRepository(pool *pgxpool.Pool) *RouteRepository {
	return &RouteRepository{pool: pool}
}

func (r *RouteRepository) FindRoute(ctx context.Context, sourceNodeID, targetNodeID int, option model.RouteOption) (*model.Route, error) {
	// フィルタ条件を構築
	var conditions []string
	conditions = append(conditions, fmt.Sprintf("level <= %d", option.Level))

	if !option.Stairs {
		conditions = append(conditions, "has_stairs = false")
	}
	if option.Accessible {
		conditions = append(conditions, "is_accessible = true")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// pgRouting のエッジクエリ
	edgeQuery := fmt.Sprintf(
		`SELECT id, source, target, distance AS cost, distance AS reverse_cost FROM edges_for_routing%s`,
		whereClause,
	)

	// pgr_dijkstra で経路探索（無向グラフ）
	query := fmt.Sprintf(`
		SELECT n.node_id, ST_Y(n.geom) AS lat, ST_X(n.geom) AS lng,
		       CASE WHEN n.building_id IS NOT NULL THEN true ELSE false END AS is_indoor,
		       r.agg_cost
		FROM pgr_dijkstra(
		    '%s',
		    $1, $2, false
		) AS r
		JOIN nodes n ON r.node = n.id
		ORDER BY r.seq`,
		strings.ReplaceAll(edgeQuery, "'", "''"),
	)

	rows, err := r.pool.Query(ctx, query, sourceNodeID, targetNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute route query: %w", err)
	}
	defer rows.Close()

	var route model.Route
	var lastCost float64
	for rows.Next() {
		var nodeID string
		var point model.RoutePoint
		var aggCost float64
		if err := rows.Scan(&nodeID, &point.Lat, &point.Lng, &point.IsIndoor, &aggCost); err != nil {
			return nil, fmt.Errorf("failed to scan route point: %w", err)
		}
		route.Points = append(route.Points, point)
		lastCost = aggCost
	}

	if len(route.Points) == 0 {
		return nil, nil
	}

	route.TotalDistance = lastCost
	route.TotalCost = lastCost

	return &route, nil
}
