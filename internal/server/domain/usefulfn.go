package domain

import (
	"runtime"
	"strconv"
)

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

func GetAction(depth int) string {
	// используется для получения имени вызывющей функции
	// по мотивам https://stackoverflow.com/questions/25927660/how-to-get-the-current-function-name
	pc, _, _, _ := runtime.Caller(depth)
	action := runtime.FuncForPC(pc).Name()
	return action
}
