package datadog

import (
	"strings"
)

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	return strings.HasPrefix(err.Error(), "API error 404 Not Found")
}

func IgnoreNotFound(err error) error {
	if IsNotFound(err) {
		return nil
	}

	return err
}
