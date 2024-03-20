package domain_test

import (
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/stretchr/testify/assert"
)

func TestExtractFloat64(t *testing.T) {
	type extractFloat64Result struct {
		value             float64
		isSuccessExpected bool
	}
	tests := []struct {
		name   string
		input  string
		result extractFloat64Result
	}{
		{
			"good value",
			"123.5",
			extractFloat64Result{
				123.5,
				true,
			},
		},
		{
			"good value 2",
			"123",
			extractFloat64Result{
				123,
				true,
			},
		},
		{
			"bad value",
			"123.F",
			extractFloat64Result{
				-1,
				false,
			},
		},
		{
			"good value",
			"1.8070544e+07",
			extractFloat64Result{
				18070544,
				true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := domain.ExtractFloat64(tt.input)
			assert.Equal(t, tt.result.value, value)
			assert.Equal(t, tt.result.isSuccessExpected, err == nil)
		})
	}
}

func TestExtractInt64(t *testing.T) {

	type extractInt64Result struct {
		value             int64
		isSuccessExpected bool
	}

	tests := []struct {
		name   string
		input  string
		result extractInt64Result
	}{
		{
			"good value",
			"123",
			extractInt64Result{
				123,
				true,
			},
		},
		{
			"bad value",
			"123F",
			extractInt64Result{
				-1,
				false,
			},
		},
		{
			"bad value 2",
			"123.5",
			extractInt64Result{
				-1,
				false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := domain.ExtractInt64(tt.input)
			assert.Equal(t, tt.result.value, value)
			assert.Equal(t, tt.result.isSuccessExpected, err == nil)
		})
	}
}
