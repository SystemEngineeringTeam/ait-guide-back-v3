package repository

import (
	"context"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/model"
)

// RoomRepository は部屋データへのアクセスを提供する。
type RoomRepository interface {
	FindAll(ctx context.Context, filter RoomFilter) ([]model.Room, error)
	FindByRoomID(ctx context.Context, roomID string) (*model.Room, error)
}

// RoomFilter は部屋検索のフィルタ条件。
type RoomFilter struct {
	BuildingID *int
	Floor      *int
	RoomType   *string
}
