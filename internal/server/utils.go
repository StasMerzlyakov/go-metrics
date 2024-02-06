package server

import (
	"fmt"
	"github.com/StasMerzlyakov/go-metrics/internal/storage"
	"regexp"
	"strconv"
)

func GetAllMetrics(counterStorage storage.MetricsStorage[int64],
	gaugeStorage storage.MetricsStorage[float64]) MetricModel {
	items := MetricModel{}
	for _, k := range counterStorage.Keys() {
		v, _ := counterStorage.Get(k)
		items.Items = append(items.Items, MetricsData{
			"counter",
			k,
			fmt.Sprintf("%v", v),
		})
	}

	for _, k := range gaugeStorage.Keys() {
		v, _ := gaugeStorage.Get(k)
		items.Items = append(items.Items, MetricsData{
			"counter",
			k,
			fmt.Sprintf("%v", v),
		})
	}
	return items
}

var decimalRegexp = regexp.MustCompile("^[-+]?([1-9][0-9]*|0?)([.][0-9]*)?$")

func CheckDecimal(value string) bool {
	return decimalRegexp.MatchString(value)
}

var integerRegexp = regexp.MustCompile("^[-+]?([1-9][0-9]*|0?)$")

func CheckInteger(value string) bool {
	return integerRegexp.MatchString(value)
}

var nameRegexp = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_]*$")

func CheckName(value string) bool {
	return nameRegexp.MatchString(value)
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
