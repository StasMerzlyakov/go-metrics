package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/app"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCheckName(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockStorage(ctrl)

	mc := app.NewMetrics(m)

	testCases := []struct {
		name   string
		input  string
		result bool
	}{
		{
			"TestCheckName_1",
			"a1s_asd1_1",
			true,
		},
		{
			"TestCheckName_2",
			"00.123",
			false,
		},
		{
			"TestCheckName_3",
			"0asd",
			false,
		},
		{
			"TestCheckName_4",
			"-A",
			false,
		},
		{
			"TestCheckName_5",
			"_asd",
			false,
		},
		{
			"TestCheckName_6",
			"A",
			true,
		},
		{
			"TestCheckName_7",
			"A0_123",
			true,
		},
		{
			"TestCheckName_8",
			"B123.1",
			false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.result, mc.CheckName(test.input))
		})
	}
}

func TestCheckMetrics(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockStorage(ctrl)

	mc := app.NewMetrics(m)

	testCases := []struct {
		name  string
		input *domain.Metrics
		isOk  bool
	}{
		{
			"CheckMetrics_1",
			&domain.Metrics{
				ID:    "a1s_asd1_1",
				MType: domain.CounterType,
				Delta: domain.DeltaPtr(1),
			},
			true,
		},
		{
			"CheckMetrics_2",
			&domain.Metrics{
				ID:    "a1s_asd1_1",
				MType: domain.GaugeType,
				Delta: domain.DeltaPtr(1),
			},
			false,
		},
		{
			"CheckMetrics_3",
			&domain.Metrics{
				ID:    "0asd",
				MType: domain.CounterType,
				Delta: domain.DeltaPtr(1),
			},
			false,
		},
		{
			"CheckMetrics_4",
			&domain.Metrics{
				ID:    "OK",
				MType: domain.GaugeType,
				Value: domain.ValuePtr(1),
			},
			true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			err := mc.CheckMetrics(test.input)
			if test.isOk {
				assert.NoError(t, err)
			} else {
				assert.True(t, errors.Is(err, domain.ErrDataFormat))
			}
		})
	}
}

func TestUpdateMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockStorage(ctrl)

	m.EXPECT().SetMetrics(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, ms []domain.Metrics) error {
			for _, m := range ms {
				assert.Equal(t, domain.GaugeType, m.MType)
			}
			return nil
		}).AnyTimes()

	m.EXPECT().AddMetrics(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, ms []domain.Metrics) error {
			for _, m := range ms {
				assert.Equal(t, domain.CounterType, m.MType)
			}
			return nil
		}).AnyTimes()

	mc := app.NewMetrics(m)
	testCases := []struct {
		name  string
		input []domain.Metrics
	}{
		{
			"add counters",
			[]domain.Metrics{
				{
					ID:    "Counter",
					MType: domain.CounterType,
					Delta: domain.DeltaPtr(1),
				},
			},
		},
		{
			"add gague",
			[]domain.Metrics{
				{
					ID:    "Gague",
					MType: domain.GaugeType,
					Value: domain.ValuePtr(1.),
				},
			},
		},
		{
			"mix",
			[]domain.Metrics{
				{
					ID:    "Gague",
					MType: domain.GaugeType,
					Value: domain.ValuePtr(1.),
				},
				{
					ID:    "Counter",
					MType: domain.CounterType,
					Delta: domain.DeltaPtr(1),
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			err := mc.UpdateAll(context.TODO(), test.input)
			assert.NoError(t, err)
		})
	}
}

func TestUpdate_Counter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockStorage(ctrl)

	input := &domain.Metrics{
		ID:    "Counter",
		MType: domain.CounterType,
		Delta: domain.DeltaPtr(1),
	}

	m.EXPECT().Add(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, ms *domain.Metrics) error {
			assert.NotNil(t, ms)
			assert.Equal(t, domain.CounterType, ms.MType)
			return nil
		}).MaxTimes(1)

	m.EXPECT().Get(gomock.Any(), gomock.Eq(input.ID), gomock.Eq(input.MType)).Return(
		input, nil,
	).MaxTimes(1)

	mc := app.NewMetrics(m)

	mtrs, err := mc.Update(context.TODO(), input)
	assert.NoError(t, err)
	assert.Equal(t, *input, *mtrs)
}

func TestUpdate_Gauge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockStorage(ctrl)

	input := &domain.Metrics{
		ID:    "Gague",
		MType: domain.GaugeType,
		Value: domain.ValuePtr(1.),
	}

	m.EXPECT().Set(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, ms *domain.Metrics) error {
			assert.NotNil(t, ms)
			assert.Equal(t, domain.GaugeType, ms.MType)
			return nil
		}).MaxTimes(1)

	m.EXPECT().Get(gomock.Any(), gomock.Eq(input.ID), gomock.Eq(input.MType)).Return(
		input, nil,
	).MaxTimes(1)

	mc := app.NewMetrics(m)

	mtrs, err := mc.Update(context.TODO(), input)
	assert.NoError(t, err)
	assert.Equal(t, *input, *mtrs)
}
