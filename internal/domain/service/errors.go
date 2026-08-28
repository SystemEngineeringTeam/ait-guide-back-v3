// Package service はドメインサービスを提供する。
package service

import "errors"

var (
	ErrNodeNotFound     = errors.New("NODE_NOT_FOUND")
	ErrNoRouteFound     = errors.New("NO_ROUTE_FOUND")
	ErrBuildingNotFound = errors.New("BUILDING_NOT_FOUND")
	ErrRoomNotFound     = errors.New("ROOM_NOT_FOUND")
	ErrInvalidParams    = errors.New("INVALID_PARAMETERS")
)
