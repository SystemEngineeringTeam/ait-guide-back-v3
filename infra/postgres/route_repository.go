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

func (r *RouteRepository) FindRoute(ctx context.Context, sourceNodeID, targetNodeID int, option model.RouteOption, weight model.CostWeight) (*model.Route, error) {
	// フィルタ条件を構築（物理的制約のみ）
	var conditions []string

	if option.Accessible {
		// バリアフリー: 階段は物理的に通れないため除外
		conditions = append(conditions, "has_stairs = false")
		conditions = append(conditions, "is_accessible = true")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// 重み付きコスト式: distance × w_level × w_stairs
	costExpr := fmt.Sprintf(
		`distance * (CASE level %s END) * (CASE WHEN has_stairs THEN %f ELSE %f END)`,
		buildLevelCase(weight.LevelWeights),
		weight.StairsWeight,
		weight.NoStairsWeight,
	)

	// pgRouting のエッジクエリ
	edgeQuery := fmt.Sprintf(
		`SELECT id, source, target, %s AS cost, %s AS reverse_cost FROM edges_for_routing%s`,
		costExpr, costExpr, whereClause,
	)

	// pgr_dijkstra で経路探索（無向グラフ）
	// edge列でedges_for_routingをJOINし、距離とコストを別々に取得
	query := fmt.Sprintf(`
		SELECT n.node_id, ST_Y(n.geom) AS lat, ST_X(n.geom) AS lng,
				CASE WHEN n.building_id IS NOT NULL THEN true ELSE false END AS is_indoor,
				r.agg_cost,
				COALESCE(efr.distance, 0) AS edge_distance
		FROM pgr_dijkstra(
			'%s'::text,
			$1::bigint, $2::bigint, false
		) AS r
		JOIN nodes n ON r.node = n.id
		LEFT JOIN edges_for_routing efr ON r.edge = efr.id
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
	var totalDistance float64
	for rows.Next() {
		var nodeID string
		var point model.RoutePoint
		var aggCost float64
		var edgeDistance float64
		if err := rows.Scan(&nodeID, &point.Lat, &point.Lng, &point.IsIndoor, &aggCost, &edgeDistance); err != nil {
			return nil, fmt.Errorf("failed to scan route point: %w", err)
		}
		route.Points = append(route.Points, point)
		lastCost = aggCost
		totalDistance += edgeDistance
	}

	if len(route.Points) == 0 {
		return nil, nil
	}

	route.TotalDistance = totalDistance
	route.TotalCost = lastCost

	return &route, nil
}

// buildLevelCase はレベル別重みのSQL CASE式を生成する。
func buildLevelCase(weights [5]float64) string {
	var b strings.Builder
	for i, w := range weights {
		fmt.Fprintf(&b, " WHEN %d THEN %f", i+1, w)
	}
	b.WriteString(" ELSE 1")
	return b.String()
}
