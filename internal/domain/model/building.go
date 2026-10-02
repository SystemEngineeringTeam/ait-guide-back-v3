// Package model はドメインのエンティティと値オブジェクトを定義する。
package model

// Building は建物エンティティ。
type Building struct {
	ID          int // buildings.building_no。nodes/rooms/photos.building_idおよびAPIのbuilding_idが参照する値
	Key         *string
	Name        string
	Description *string
	Affiliation *string
	Photos      []Photo
}
