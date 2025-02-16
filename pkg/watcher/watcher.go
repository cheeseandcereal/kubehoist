/*
Copyright 2025.

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

package watcher

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"

	controllerv1alpha1 "github.com/cheeseandcereal/kubehoist/api/v1alpha1"
)

// GenericWatcher watches an arbitrary resource
type GenericWatcher struct {
	client.Client
	GVK             schema.GroupVersionKind
	ControllerWatch client.ObjectKey
}

func (g *GenericWatcher) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the partial arbitray resource
	obj := &metav1.PartialObjectMetadata{}
	obj.SetGroupVersionKind(g.GVK)
	if err := g.Get(ctx, req.NamespacedName, obj); err != nil {
		log.Error(err, "ignoring not found error for custom resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Get the corresponding controller watch
	controllerWatch := &controllerv1alpha1.ControllerWatch{}
	if err := g.Get(ctx, g.ControllerWatch, controllerWatch); err != nil {
		log.Error(err, "unable to fetch corresponding ControllerWatch resource for custom resource", "gvk", g.GVK, "ControllerWatch", g.ControllerWatch)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("usage of watched custom resource detected", "gvk", g.GVK, "ControllerWatch", g.ControllerWatch)

	if controllerWatch.Status.ControllerInstallationStatus != controllerv1alpha1.ControllerInstallationStatusInstalled &&
		controllerWatch.Status.ControllerInstallationStatus != controllerv1alpha1.ControllerInstallationStatusPending {
		// Set the controller watch status to pending to trigger the installation
		log.Info("updating controller watch controller installation status to pending", "ControllerWatch", g.ControllerWatch)
		controllerWatch.Status.ControllerInstallationStatus = controllerv1alpha1.ControllerInstallationStatusPending
		controllerWatch.Status.LastUpdated = &metav1.Time{Time: time.Now()}
		if err := g.Status().Update(ctx, controllerWatch); err != nil {
			log.Error(err, "could not update ControllerWatch status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller to start the reconcile watch loop with manager
func (g *GenericWatcher) SetupWithManager(mgr ctrl.Manager) error {
	obj := &metav1.PartialObjectMetadata{}
	obj.SetGroupVersionKind(g.GVK)
	return ctrl.NewControllerManagedBy(mgr).
		WatchesMetadata(obj, &handler.EnqueueRequestForObject{}).
		Named(g.GVK.String()).
		Complete(g)
}
