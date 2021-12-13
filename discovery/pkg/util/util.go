package util

import (
	"strconv"
)

func ConvertStringToUint(val string) uint64 {
	ret, _ := strconv.ParseUint(val, 10, 64)
	return ret
}

func ConvertUnitToString(val uint64) string {
	return strconv.FormatUint(val, 10)
}
