package postgres

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func NewStorage(logger *zap.SugaredLogger, databaseUrl string) *storage {
	return &storage{
		databaseUrl: databaseUrl,
		logger:      logger,
	}
}

type storage struct {
	db          *sql.DB
	databaseUrl string
	logger      *zap.SugaredLogger
}

func (st *storage) Start(ctx context.Context) error {

	if db, err := sql.Open("pgx", st.databaseUrl); err != nil {
		st.logger.Infow("Start", "status", "error", "msg", err.Error())
		return err
	} else {
		st.logger.Infow("Start", "status", "ok")
		st.db = db
		return nil
	}
}

func (st *storage) Ping(ctx context.Context) error {
	if err := st.db.PingContext(ctx); err != nil {
		st.logger.Infow("Ping", "status", "error", "msg", err.Error())
		return err
	} else {
		st.logger.Infow("Ping", "status", "ok")
		return nil
	}
}

func (st *storage) Stop(ctx context.Context) error {
	if err := st.db.Close(); err != nil {
		st.logger.Infow("Stop", "status", "error", "msg", err.Error())
		return err
	} else {
		st.logger.Infow("Stop", "status", "ok")
		return nil
	}
}
