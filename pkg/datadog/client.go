package datadog

import (
	"os"

	"github.com/zorkian/go-datadog-api"
)

type Client = datadog.Client

func NewClient() *datadog.Client {
	apiKey := os.Getenv("DD_API_KEY")
	appKey := os.Getenv("DD_APPLICATION_KEY")

	return datadog.NewClient(apiKey, appKey)
}
