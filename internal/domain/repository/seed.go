package repository

import "context"

// SeedRepository はデータの初期投入を提供する。
type SeedRepository interface {
	SeedBuildings(ctx context.Context, path string) error
	SeedNodes(ctx context.Context, path string) error
	SeedEdges(ctx context.Context, path string) error
	SeedRooms(ctx context.Context, path string) error
	SeedPhotos(ctx context.Context, path string) error
}
