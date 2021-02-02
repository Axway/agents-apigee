package util

import (
	"fmt"
	"net/url"
	"strconv"
)

// IsValidURL - tests a string to determine if it is a well-structured url or not.
func IsValidURL(testString string) bool {
	_, err := url.ParseRequestURI(testString)
	if err != nil {
		return false
	}

	u, err := url.Parse(testString)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func ConvertStringToUint(val string) uint64 {
	ret, _ := strconv.ParseUint(val, 10, 64)
	return ret
}

func ConvertUnitToString(val uint64) string {
	return strconv.FormatUint(val, 10)
}

func FormatRemoteAPIID(proxyName, deployedEnvName, revisionName string) string {
	return fmt.Sprintf("%v-%v-%v", proxyName, deployedEnvName, revisionName)
}
