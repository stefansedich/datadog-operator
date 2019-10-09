package datadog

import (
	"strings"
)

func IsNotFound(err error) bool {
	return strings.HasPrefix(err.Error(), "API error 404 Not Found")
}
