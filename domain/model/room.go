package model

// Room は部屋エンティティ。
type Room struct {
	ID          int
	RoomID      string
	BuildingID  int
	NodeID      *string
	Name        *string
	Description *string
	Floor       *int
	Capacity    *int
	RoomType    *string
}
