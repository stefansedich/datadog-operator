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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitoringv1alpha1 "github.com/stefansedich/datadog-operator/api/v1alpha1"
	"github.com/stefansedich/datadog-operator/internal/controller"
	"github.com/stefansedich/datadog-operator/internal/datadog"
)

const (
	finalizerName = "monitoring.datadog.com.monitor"
)

// MonitorReconciler reconciles a Monitor object
type MonitorReconciler struct {
	client.Client
	Log           logr.Logger
	DataDogClient *datadog.Client
}

func isBeingCreated(monitor *monitoringv1alpha1.Monitor) bool {
	return monitor.Status.MonitorID == 0
}

func isBeingDeleted(monitor *monitoringv1alpha1.Monitor) bool {
	return !monitor.ObjectMeta.DeletionTimestamp.IsZero() &&
		controller.HasFinalizer(&monitor.ObjectMeta, finalizerName)
}

func createMonitor(reconciler *MonitorReconciler, monitor *monitoringv1alpha1.Monitor) error {
	client := reconciler.DataDogClient
	log := reconciler.Log

	ddMonitor := &datadog.Monitor{}
	_, err := datadog.ChangeMonitor(ddMonitor, monitor)
	if err != nil {
		return err
	}

	newDDMonitor, err := client.CreateMonitor(ddMonitor)
	if err != nil {
		return err
	}

	monitor.Status.MonitorID = *newDDMonitor.Id

	err = reconciler.Status().Update(context.Background(), monitor)
	if err != nil {
		return err
	}

	controller.AddFinalizer(&monitor.ObjectMeta, finalizerName)

	err = reconciler.Update(context.Background(), monitor)
	if err != nil {
		return err
	}

	log.V(1).Info("Successfully created monitor", "monitor_id", *newDDMonitor.Id)

	return nil
}

func updateMonitor(reconciler *MonitorReconciler, monitor *monitoringv1alpha1.Monitor) error {
	client := reconciler.DataDogClient
	log := reconciler.Log.WithValues("monitor_id", monitor.Status.MonitorID)

	ddMonitor, err := client.GetMonitor(monitor.Status.MonitorID)
	if err != nil {
		if datadog.IsNotFound(err) {
			log.V(1).Info("Existing monitor not found, creating again")

			return createMonitor(reconciler, monitor)
		}

		return err
	}

	changed, err := datadog.ChangeMonitor(ddMonitor, monitor)
	if err != nil {
		return err
	}

	if !changed {
		log.V(1).Info("Skipping update of unchanged monitor")

		return nil
	}

	err = client.UpdateMonitor(ddMonitor)
	if err != nil {

		return err
	}

	log.V(1).Info("Successfully updated monitor")

	return nil
}

func deleteMonitor(reconciler *MonitorReconciler, monitor *monitoringv1alpha1.Monitor) error {
	client := reconciler.DataDogClient
	log := reconciler.Log.WithValues("monitor_id", monitor.Status.MonitorID)

	err := client.DeleteMonitor(monitor.Status.MonitorID)
	if err != nil {
		return datadog.IgnoreNotFound(err)
	}

	controller.RemoveFinalizer(&monitor.ObjectMeta, finalizerName)

	err = reconciler.Update(context.Background(), monitor)
	if err != nil {
		return err
	}

	log.V(1).Info("Successfully deleted monitor")

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
		return ctrl.Result{}, controller.IgnoreNotFound(err)
	}

	if isBeingDeleted(monitor) {
		err := deleteMonitor(r, monitor)
		if err != nil {
			log.Error(err, "Failed to delete monitor")

			return ctrl.Result{}, err
		}
	} else if isBeingCreated(monitor) {
		err := createMonitor(r, monitor)
		if err != nil {
			log.Error(err, "Failed to create monitor")

			return ctrl.Result{}, err
		}
	} else {
		err := updateMonitor(r, monitor)
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
