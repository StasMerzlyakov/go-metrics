package domain

import "strconv"

func DeltaPtr(v int64) *int64 {
	return &v
}

func ValuePtr(v float64) *float64 {
	return &v
}

func ExtractFloat64(valueStr string) (float64, error) {
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return -1, err
	}
	return value, nil
}

func ExtractInt64(valueStr string) (int64, error) {
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return -1, err
	}
	return value, nil
}
