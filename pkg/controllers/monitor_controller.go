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
	"github.com/stefansedich/datadog-operator/pkg/datadog"
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
		hasFinalizer(&monitor.ObjectMeta, finalizerName)
}

func (r *MonitorReconciler) createMonitor(req ctrl.Request, monitor *monitoringv1alpha1.Monitor) error {
	client := r.DataDogClient
	log := r.Log.WithValues("monitor", req.NamespacedName)

	log.Info("Creating monitor")

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

	err = r.Status().Update(context.Background(), monitor)
	if err != nil {
		return err
	}

	addFinalizer(&monitor.ObjectMeta, finalizerName)

	err = r.Update(context.Background(), monitor)
	if err != nil {
		return err
	}

	log.Info("Successfully created monitor", "monitor_id", *newDDMonitor.Id)

	return nil
}

func (r *MonitorReconciler) updateMonitor(req ctrl.Request, monitor *monitoringv1alpha1.Monitor) error {
	client := r.DataDogClient
	log := r.Log.WithValues(
		"monitor",
		req.NamespacedName,
		"monitor_id",
		monitor.Status.MonitorID,
	)

	log.Info("Updating monitor")

	ddMonitor, err := client.GetMonitor(monitor.Status.MonitorID)
	if err != nil {
		if datadog.IsNotFound(err) {
			log.Info("Existing monitor not found, creating again")

			return r.createMonitor(req, monitor)
		}

		return err
	}

	changed, err := datadog.ChangeMonitor(ddMonitor, monitor)
	if err != nil {
		return err
	}

	if !changed {
		log.Info("Skipping update of unchanged monitor")

		return nil
	}

	err = client.UpdateMonitor(ddMonitor)
	if err != nil {
		return err
	}

	log.Info("Successfully updated monitor")

	return nil
}

func (r *MonitorReconciler) deleteMonitor(req ctrl.Request, monitor *monitoringv1alpha1.Monitor) error {
	client := r.DataDogClient
	log := r.Log.WithValues(
		"monitor",
		req.NamespacedName,
		"monitor_id",
		monitor.Status.MonitorID,
	)

	log.Info("Deleting monitor")

	err := client.DeleteMonitor(monitor.Status.MonitorID)
	if err != nil {
		return datadog.IgnoreNotFound(err)
	}

	removeFinalizer(&monitor.ObjectMeta, finalizerName)

	err = r.Update(context.Background(), monitor)
	if err != nil {
		return err
	}

	log.Info("Successfully deleted monitor")

	return nil
}

func (r *MonitorReconciler) handleError(req ctrl.Request, err error) (ctrl.Result, error) {
	log := r.Log.WithValues("monitor", req.NamespacedName)

	if datadog.IsBadRequest(err) {
		// TODO: Extract reason from error
		log.Error(err, "Bad request to DataDog API")

		return ctrl.Result{}, nil
	} else if datadog.IsForbidden(err) {
		log.Error(nil, "Failed to authenticate with DataDog API")

		return ctrl.Result{}, nil
	} else {
		return ctrl.Result{}, err
	}
}

// +kubebuilder:rbac:groups=monitoring.datadog.com,resources=monitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.datadog.com,resources=monitors/status,verbs=get;update;patch

func (r *MonitorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	monitor := &monitoringv1alpha1.Monitor{}
	err := r.Get(ctx, req.NamespacedName, monitor)
	if err != nil {
		return ctrl.Result{}, ignoreNotFound(err)
	}

	if isBeingDeleted(monitor) {
		err := r.deleteMonitor(req, monitor)
		if err != nil {
			return r.handleError(req, err)
		}
	} else if isBeingCreated(monitor) {
		err := r.createMonitor(req, monitor)
		if err != nil {
			return r.handleError(req, err)
		}
	} else {
		err := r.updateMonitor(req, monitor)
		if err != nil {
			return r.handleError(req, err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *MonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1alpha1.Monitor{}).
		Complete(r)
}
