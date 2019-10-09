package datadog

import (
	"strings"
)

func IsBadRequest(err error) bool {
	if err == nil {
		return false
	}

	return strings.HasPrefix(err.Error(), "API error 400 Bad Request")
}

func IsForbidden(err error) bool {
	if err == nil {
		return false
	}

	return strings.HasPrefix(err.Error(), "API error 403 Forbidden")
}

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
