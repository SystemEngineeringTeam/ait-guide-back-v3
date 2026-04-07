package service

import "github.com/SystemEngineeringTeam/ait-guide-back-v3/domain/model"

// CostWeightFromOption は検索条件に応じたコスト重みを返す。
func CostWeightFromOption(option model.RouteOption) model.CostWeight {
	return model.CostWeight{
		LevelWeights:   levelWeights(option),
		StairsWeight:   stairsWeight(option),
		NoStairsWeight: 1,
	}
}

// levelWeights は検索条件に応じたレベル別重みを返す。
//
// level パラメータの意味:
//
//	0: 距離のみ（全レベル同価、重みなし）
//	1: 大通り優先（脇道に大きなペナルティ）
//	2: やや大通り優先
//	3: デフォルト（バランス型）
//	4: やや脇道優先（大通りにペナルティ）
//	5: 脇道優先（大通りに大きなペナルティ）
func levelWeights(option model.RouteOption) [5]float64 {
	//                      L1    L2    L3    L4    L5
	switch option.Level {
	case 0:
		return [5]float64{1, 1, 1, 1, 1}
	case 1:
		return [5]float64{1, 5, 20, 50, 100}
	case 2:
		return [5]float64{1, 2, 5, 10, 20}
	case 4:
		return [5]float64{20, 10, 5, 2, 1}
	case 5:
		return [5]float64{100, 50, 20, 5, 1}
	default:
		// level=3（デフォルト）
		return [5]float64{1, 2, 5, 10, 20}
	}
}

// stairsWeight は検索条件に応じた階段重みを返す。
func stairsWeight(option model.RouteOption) float64 {
	if option.Level == 0 {
		// 距離のみ: 全経路同価
		return 1
	}
	if option.Accessible {
		// バリアフリー: 階段はハードフィルタで除外済み（物理的に通れない）
		return 1
	}
	if !option.Stairs {
		// 階段拒否: 大きなペナルティ（通れるが避けたい）
		return 10
	}
	// 階段許可: 軽いペナルティ
	return 1.5
}
