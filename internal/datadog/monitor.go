package datadog

import (
	"encoding/json"

	datadogapi "github.com/zorkian/go-datadog-api"

	monitoringv1alpha1 "github.com/stefansedich/datadog-operator/api/v1alpha1"
)

type Monitor = datadogapi.Monitor
type Options = datadogapi.Options

func HasMonitorChanged(ddMonitor *Monitor, monitor *monitoringv1alpha1.Monitor) bool {
	return true
}

func PopulateMonitor(ddMonitor *Monitor, monitor *monitoringv1alpha1.Monitor) error {
	spec := monitor.Spec
	status := monitor.Status
	options := &Options{}

	err := json.Unmarshal(spec.Options, options)
	if err != nil {
		return err
	}

	ddMonitor.Id = &status.MonitorID
	ddMonitor.Type = &spec.Type
	ddMonitor.Name = &spec.Name
	ddMonitor.Message = &spec.Message
	ddMonitor.Query = &spec.Query
	ddMonitor.Tags = spec.Tags
	ddMonitor.Options = options

	return nil
}
