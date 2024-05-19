package app

import (
	"context"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

func NewAdminApp(pinger Pinger) *adminApp {
	return &adminApp{
		pinger: pinger,
	}
}

type adminApp struct {
	pinger Pinger
}

func (admApp *adminApp) Ping(ctx context.Context) error {
	return admApp.pinger.Ping(ctx)
}
