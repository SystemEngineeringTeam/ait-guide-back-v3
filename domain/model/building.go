package model

// Building は建物エンティティ。
type Building struct {
	ID          int
	Name        string
	Description *string
	Affiliation *string
	Photos      []Photo
}
