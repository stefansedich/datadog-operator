package datadog

import (
	"encoding/json"

	"github.com/mitchellh/hashstructure"
	datadog "github.com/zorkian/go-datadog-api"

	monitoringv1alpha1 "github.com/stefansedich/datadog-operator/api/v1alpha1"
)

type Monitor = datadog.Monitor
type Options = datadog.Options

func ChangeMonitor(ddMonitor *Monitor, monitor *monitoringv1alpha1.Monitor) (bool, error) {
	spec := monitor.Spec

	originalHash, err := hashstructure.Hash(ddMonitor, nil)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(spec.Options.Raw, &ddMonitor.Options)
	if err != nil {
		return false, err
	}

	ddMonitor.Id = &monitor.Status.MonitorID
	ddMonitor.Type = &spec.Type
	ddMonitor.Name = &spec.Name
	ddMonitor.Message = &spec.Message
	ddMonitor.Query = &spec.Query
	ddMonitor.Tags = spec.Tags

	newHash, err := hashstructure.Hash(ddMonitor, nil)
	if err != nil {
		return false, err
	}

	return originalHash != newHash, nil
}
