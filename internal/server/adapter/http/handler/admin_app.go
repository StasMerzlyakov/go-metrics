package handler

import (
	"context"
)

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . AdminApp
type AdminApp interface {
	Ping(ctx context.Context) error
}
