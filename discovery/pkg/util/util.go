package util

import (
	"net/url"
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
