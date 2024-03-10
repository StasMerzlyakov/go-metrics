package postgres

import (
	"context"
	"database/sql"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func NewStorage(logger *zap.SugaredLogger, databaseURL string) *storage {
	return &storage{
		databaseURL: databaseURL,
		logger:      logger,
	}
}

type storage struct {
	db          *sql.DB
	databaseURL string
	logger      *zap.SugaredLogger
}

func (st *storage) SetAllMetrics(ctx context.Context, in []domain.Metrics) error {
	panic("SetAllMetrics not implemented")
}

func (st *storage) GetAllMetrics(ctx context.Context) ([]domain.Metrics, error) {
	panic("GetAllMetrics not implemented")
}

func (st *storage) Set(ctx context.Context, m *domain.Metrics) error {
	panic("Set not implemented")
}

func (st *storage) Add(ctx context.Context, m *domain.Metrics) error {
	panic("Add not implemented")
}
func (st *storage) Get(ctx context.Context, id string, mType domain.MetricType) (*domain.Metrics, error) {
	panic("Get not implemented")
}

func (st *storage) Start(ctx context.Context) error {

	// TODO - create table if not exists
	if db, err := sql.Open("pgx", st.databaseURL); err != nil {
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
