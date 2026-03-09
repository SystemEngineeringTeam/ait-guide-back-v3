package model

// Photo は建物写真エンティティ。
type Photo struct {
	ID           int
	BuildingID   int
	URL          string
	Caption      *string
	DisplayOrder int
}
