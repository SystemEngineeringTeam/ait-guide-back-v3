package model

// NodeType はノード種別を表す。
type NodeType string

const (
	NodeTypeEntrance NodeType = "entrance"
	NodeTypeRoad     NodeType = "road"
	NodeTypeDoor     NodeType = "door"
)

// Node は経路ノードエンティティ。
type Node struct {
	ID         int
	NodeID     string
	Lat        float64
	Lng        float64
	NodeType   NodeType
	BuildingID *int
	Floor      *int
}
