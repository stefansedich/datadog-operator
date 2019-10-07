/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitoringv1alpha1 "github.com/stefansedich/datadog-operator/api/v1alpha1"
	"github.com/stefansedich/datadog-operator/internal/datadog"
)

// MonitorReconciler reconciles a Monitor object
type MonitorReconciler struct {
	client.Client
	Log logr.Logger
}

func createMonitor(reconciler *MonitorReconciler, log logr.Logger, monitor *monitoringv1alpha1.Monitor) error {
	client := datadog.NewClient()
	ctx := context.Background()

	ddMonitor := &datadog.Monitor{}
	err := datadog.PopulateMonitor(ddMonitor, monitor)
	if err != nil {
		return err
	}

	newDDMonitor, err := client.CreateMonitor(ddMonitor)
	if err != nil {
		return err
	}

	monitor.Status.MonitorID = *newDDMonitor.Id

	err = reconciler.Status().Update(ctx, monitor)
	if err != nil {
		log.Error(err, "Failed to update monitor status")

		return err
	}

	log.V(1).Info("Successfully created monitor", "monitor_id", *newDDMonitor.Id)

	return nil
}

func updateMonitor(log logr.Logger, monitor *monitoringv1alpha1.Monitor) error {
	client := datadog.NewClient()

	ddMonitor, err := client.GetMonitor(monitor.Status.MonitorID)
	if err != nil {
		return err
	}

	if !datadog.HasMonitorChanged(ddMonitor, monitor) {
		log.V(1).Info("Skipping updating unchanged Monitor")

		return nil
	}

	err = datadog.PopulateMonitor(ddMonitor, monitor)
	if err != nil {
		return err
	}

	err = client.UpdateMonitor(ddMonitor)
	if err != nil {

		return err
	}

	log.V(1).Info("Successfully updated monitor")

	return nil
}

// +kubebuilder:rbac:groups=monitoring.datadog.com,resources=monitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.datadog.com,resources=monitors/status,verbs=get;update;patch

func (r *MonitorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("monitor", req.NamespacedName)

	monitor := &monitoringv1alpha1.Monitor{}
	err := r.Get(ctx, req.NamespacedName, monitor)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	log = log.WithValues(
		"monitor_id",
		monitor.Status.MonitorID,
		"monitor_name",
		monitor.Spec.Name,
	)

	// TODO: Handle delete

	if monitor.Status.MonitorID == 0 {
		err := createMonitor(r, log, monitor)
		if err != nil {
			log.Error(err, "Failed to create monitor")

			return ctrl.Result{}, err
		}
	} else {
		err = updateMonitor(log, monitor)
		if err != nil {
			log.Error(err, "Failed to update monitor")

			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *MonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1alpha1.Monitor{}).
		Complete(r)
}
