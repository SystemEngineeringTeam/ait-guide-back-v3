package usecase

import (
	"context"

	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/model"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/repository"
	"github.com/SystemEngineeringTeam/ait-guide-back-v3/internal/domain/service"
)

// RoomUsecase は部屋に関するユースケース。
type RoomUsecase struct {
	repo repository.RoomRepository
}

// NewRoomUsecase は新しいRoomUsecaseを作成する。
func NewRoomUsecase(repo repository.RoomRepository) *RoomUsecase {
	return &RoomUsecase{repo: repo}
}

// List は部屋一覧を取得する。
func (u *RoomUsecase) List(ctx context.Context, filter repository.RoomFilter) ([]model.Room, error) {
	return u.repo.FindAll(ctx, filter)
}

// GetByRoomID は部屋情報を取得する。
func (u *RoomUsecase) GetByRoomID(ctx context.Context, roomID string) (*model.Room, error) {
	rm, err := u.repo.FindByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if rm == nil {
		return nil, service.ErrRoomNotFound
	}
	return rm, nil
}
