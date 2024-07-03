package app

import (
	"context"
)

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
