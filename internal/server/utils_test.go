package server

import (
	"testing"

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
			value, err := ExtractFloat64(tt.input)
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
			value, err := ExtractInt64(tt.input)
			assert.Equal(t, tt.result.value, value)
			assert.Equal(t, tt.result.isSuccessExpected, err == nil)
		})
	}
}

func TestCheckName(t *testing.T) {
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
			assert.Equal(t, test.result, CheckName(test.input))
		})
	}
}
