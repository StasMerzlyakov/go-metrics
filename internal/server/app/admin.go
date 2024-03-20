package app

import (
	"context"

	"go.uber.org/zap"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

func NewAdminApp(suga *zap.SugaredLogger, pinger Pinger) *adminApp {
	return &adminApp{
		suga:   suga,
		pinger: pinger,
	}
}

type adminApp struct {
	suga   *zap.SugaredLogger
	pinger Pinger
}

func (admApp *adminApp) Ping(ctx context.Context) error {
	return admApp.pinger.Ping(ctx)
}
