package model

// RoutePoint は経路上の一点を表す。
type RoutePoint struct {
	Lat      float64
	Lng      float64
	IsIndoor bool
}

// Route は探索結果の経路を表す。
type Route struct {
	Points        []RoutePoint
	TotalDistance  float64
	TotalCost     float64
}

// RouteOption は経路探索のフィルタリングオプション。
type RouteOption struct {
	Level      int
	Stairs     bool
	Accessible bool
	Indoor     bool
}
