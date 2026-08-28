package model

// CostWeight は経路コスト計算の重みを定義する。
// エッジコスト = distance × LevelWeight × StairsWeight
type CostWeight struct {
	// LevelWeights はレベルごとの重み（index 0=level1, 4=level5）。
	LevelWeights [5]float64
	// StairsWeight は階段ありエッジの重み。
	StairsWeight float64
	// NoStairsWeight は階段なしエッジの重み。
	NoStairsWeight float64
}
