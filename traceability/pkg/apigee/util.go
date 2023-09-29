package apigee

import "strconv"

func parseFloatToFloat64(value string) float64 {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return float64(0)
	}
	return f
}

func parseFloatToInt64(value string) int64 {
	return int64(parseFloatToFloat64(value))
}
