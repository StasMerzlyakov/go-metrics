package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"

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
	sm          sync.Mutex
}

var createCounterTableSQL = `CREATE TABLE IF NOT EXISTS counter(
	name text not null,
	value bigint,
	PRIMARY KEY(name)
);`

var createGaugeTableSQL = `CREATE TABLE IF NOT EXISTS gauge(
	name text not null,
	value double precision,
	PRIMARY KEY(name)
);`

func (st *storage) SetAllMetrics(ctx context.Context, in []domain.Metrics) error {

	if err := st.initIfNessessary(ctx); err != nil {
		return err
	}
	_, err := st.db.ExecContext(ctx, "TRUNCATE counter,gauge")

	if err != nil {
		return err
	}

	var gaugeList []gauge
	var counterList []counter

	for _, metrics := range in {
		switch metrics.MType {
		case domain.CounterType:
			counter := counter{
				name:  metrics.ID,
				value: *metrics.Delta,
			}
			counterList = append(counterList, counter)
		case domain.GaugeType:
			gauge := gauge{
				name:  metrics.ID,
				value: *metrics.Value,
			}
			gaugeList = append(gaugeList, gauge)
		default:
			return fmt.Errorf("unknown MType %v", metrics.MType)
		}
	}

	if err := st.insertCounterList(ctx, counterList); err != nil {
		return err
	}

	if err := st.insertGaugeList(ctx, gaugeList); err != nil {
		return err
	}

	return nil
}

func (st *storage) GetAllMetrics(ctx context.Context) ([]domain.Metrics, error) {

	if err := st.initIfNessessary(ctx); err != nil {
		return nil, err
	}

	var metricsList []domain.Metrics
	gaugeList, err := st.getAllGauge(ctx)
	if err != nil {
		return nil, err
	}

	for _, gauge := range gaugeList {
		value := gauge.value
		metricsList = append(metricsList, domain.Metrics{
			ID:    gauge.name,
			MType: domain.GaugeType,
			Value: &value,
		})
	}

	counterList, err := st.getAllCounter(ctx)
	if err != nil {
		return nil, err
	}

	for _, counter := range counterList {
		delta := counter.value
		metricsList = append(metricsList, domain.Metrics{
			ID:    counter.name,
			MType: domain.CounterType,
			Delta: &delta,
		})
	}

	return metricsList, nil
}

func (st *storage) Set(ctx context.Context, m *domain.Metrics) error {

	if err := st.initIfNessessary(ctx); err != nil {
		return err
	}

	switch m.MType {
	case domain.CounterType:
		delta := *m.Delta
		_, err := st.db.ExecContext(ctx, "INSERT INTO counter(name, value) VALUES ($1, $2) ON CONFLICT(name) DO UPDATE SET value = EXCLUDED.value", m.ID, delta)
		return err
	case domain.GaugeType:
		value := *m.Value
		_, err := st.db.ExecContext(ctx, "INSERT INTO gauge(name, value) VALUES ($1, $2) ON CONFLICT(name) DO UPDATE SET value = EXCLUDED.value", m.ID, value)
		return err
	default:
		return fmt.Errorf("unknown MType %v", m.MType)
	}
}

func (st *storage) Add(ctx context.Context, m *domain.Metrics) error {

	if err := st.initIfNessessary(ctx); err != nil {
		return err
	}

	switch m.MType {
	case domain.CounterType:
		delta := *m.Delta
		_, err := st.db.ExecContext(ctx, "INSERT INTO counter(name, value) VALUES ($1, $2) ON CONFLICT(name) DO UPDATE SET value = counter.value + EXCLUDED.value", m.ID, delta)
		return err
	case domain.GaugeType:
		value := *m.Value
		_, err := st.db.ExecContext(ctx, "INSERT INTO gauge(name, value) VALUES ($1, $2) ON CONFLICT(name) DO UPDATE SET value = gauge.value + EXCLUDED.value", m.ID, value)
		return err

	default:
		return fmt.Errorf("unknown MType %v", m.MType)
	}
}

func (st *storage) Get(ctx context.Context, id string, mType domain.MetricType) (*domain.Metrics, error) {
	if err := st.initIfNessessary(ctx); err != nil {
		return nil, err
	}
	switch mType {
	case domain.CounterType:
		return st.getCounter(ctx, id)
	case domain.GaugeType:
		return st.getGauge(ctx, id)
	default:
		return nil, fmt.Errorf("unknown MType %v", mType)
	}
}

func (st *storage) Start(ctx context.Context) error {
	err := st.initIfNessessary(ctx)
	return err
}

func (st *storage) Ping(ctx context.Context) error {
	if err := st.initIfNessessary(ctx); err != nil {
		return err
	}

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

func (st *storage) insertCounterList(ctx context.Context, counterList []counter) error {

	if len(counterList) == 0 {
		return nil
	}

	// Попытка реализовать bulk insert средствами database/sql
	// https://stackoverflow.com/questions/12486436/how-do-i-batch-sql-statements-with-package-database-sql
	valueStrings := make([]string, 0, len(counterList))
	valueArgs := make([]interface{}, 0, len(counterList)*2)

	for i, counter := range counterList {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, counter.name)
		valueArgs = append(valueArgs, counter.value)
	}

	sqlQuery := fmt.Sprintf("INSERT INTO counter (name, value) VALUES %s", strings.Join(valueStrings, ","))

	_, err := st.db.ExecContext(ctx, sqlQuery, valueArgs...)
	return err
}

func (st *storage) insertGaugeList(ctx context.Context, gaugeList []gauge) error {

	if len(gaugeList) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(gaugeList))
	valueArgs := make([]interface{}, 0, len(gaugeList)*2)

	for i, gauge := range gaugeList {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, gauge.name)
		valueArgs = append(valueArgs, gauge.value)
	}

	sqlQuery := fmt.Sprintf("INSERT INTO gauge (name, value) VALUES %s", strings.Join(valueStrings, ","))

	_, err := st.db.ExecContext(ctx, sqlQuery, valueArgs...)
	return err
}

func (st *storage) getAllCounter(ctx context.Context) ([]counter, error) {
	var counterList []counter

	rows, err := st.db.QueryContext(ctx, "SELECT name, value from counter")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c counter
		err = rows.Scan(&c.name, &c.value)
		if err != nil {
			return nil, err
		}
		counterList = append(counterList, c)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return counterList, nil
}

func (st *storage) getAllGauge(ctx context.Context) ([]gauge, error) {

	var gaugeList []gauge

	rows, err := st.db.QueryContext(ctx, "SELECT name, value from gauge")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c gauge
		err = rows.Scan(&c.name, &c.value)
		if err != nil {
			return nil, err
		}
		gaugeList = append(gaugeList, c)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return gaugeList, nil
}

func (st *storage) getCounter(ctx context.Context, id string) (*domain.Metrics, error) {

	rows, err := st.db.QueryContext(ctx, "SELECT name, value from counter WHERE name = $1", id)
	if err != nil {
		st.logger.Infow("getCounter", "status", "error", "msg", err.Error())
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var name string
		var delta int64
		err = rows.Scan(&name, &delta)
		if err != nil {
			st.logger.Infow("getCounter", "status", "error", "msg", err.Error())
			return nil, err
		}

		return &domain.Metrics{
			ID:    name,
			MType: domain.CounterType,
			Delta: &delta,
		}, nil
	}
	return nil, nil
}

func (st *storage) getGauge(ctx context.Context, id string) (*domain.Metrics, error) {

	rows, err := st.db.QueryContext(ctx, "SELECT name, value from gauge WHERE name = $1", id)
	if err != nil {
		st.logger.Infow("getGauge", "status", "error", "msg", err.Error())
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var name string
		var value float64
		err = rows.Scan(&name, &value)
		if err != nil {
			st.logger.Infow("getGauge", "status", "error", "msg", err.Error())
			return nil, err
		}

		return &domain.Metrics{
			ID:    name,
			MType: domain.GaugeType,
			Value: &value,
		}, nil
	}
	return nil, nil
}

func (st *storage) initIfNessessary(ctx context.Context) error {
	st.sm.Lock()
	defer st.sm.Unlock()
	if st.db == nil {
		if db, err := sql.Open("pgx", st.databaseURL); err != nil {
			st.logger.Infow("InitIfNessessary", "status", "error", "msg", err.Error())
			return err
		} else {
			st.db = db

			if _, err := st.db.ExecContext(ctx, createCounterTableSQL); err != nil {
				st.logger.Infow("InitIfNessessary", "status", "error", "msg", err.Error())
				return err
			}

			if _, err := st.db.ExecContext(ctx, createGaugeTableSQL); err != nil {
				st.logger.Infow("InitIfNessessary", "status", "error", "msg", err.Error())
				return err
			}
			st.logger.Infow("InitIfNessessary", "status", "ok", "msg", "tables created")
			return nil
		}
	}
	return nil
}

type gauge struct {
	name  string
	value float64
}

type counter struct {
	name  string
	value int64
}
