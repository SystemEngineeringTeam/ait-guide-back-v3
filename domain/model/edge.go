package model

// Edge は経路エッジエンティティ。
type Edge struct {
	ID           int
	NodeIDFrom   string
	NodeIDTarget string
	Distance     float64
	Cost         *float64
	Level        int
	HasStairs    bool
	IsAccessible bool
	IsIndoor     bool
}
