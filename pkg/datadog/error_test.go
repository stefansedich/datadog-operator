package datadog_test

import (
	"fmt"
	"testing"

	"gotest.tools/assert"

	"github.com/stefansedich/datadog-operator/pkg/datadog"
)

func TestIsBadRequest(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{fmt.Errorf("foo"), false},
		{fmt.Errorf("API error 400 Bad Request: foo"), true},
	}

	for _, test := range tests {
		isBadRequest := datadog.IsBadRequest(test.err)

		assert.Equal(t, isBadRequest, test.expected)
	}
}

func TestIsForbidden(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{fmt.Errorf("foo"), false},
		{fmt.Errorf("API error 403 Forbidden: foo"), true},
	}

	for _, test := range tests {
		isForbidden := datadog.IsForbidden(test.err)

		assert.Equal(t, isForbidden, test.expected)
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{fmt.Errorf("foo"), false},
		{fmt.Errorf("API error 404 Not Found: foo"), true},
	}

	for _, test := range tests {
		isNotFound := datadog.IsNotFound(test.err)

		assert.Equal(t, isNotFound, test.expected)
	}
}

func TestIgnoreNotFound(t *testing.T) {
	tests := []struct {
		err   error
		is404 bool
	}{
		{fmt.Errorf("foo"), false},
		{fmt.Errorf("API error 404 Not Found: foo"), true},
	}

	for _, test := range tests {
		err := datadog.IgnoreNotFound(test.err)

		if test.is404 {
			assert.NilError(t, err)
		} else {
			assert.Equal(t, err, test.err)
		}
	}
}
