package handler

import "context"

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . AdminOperation
type AdminOperation interface {
	Ping(ctx context.Context) error
}
