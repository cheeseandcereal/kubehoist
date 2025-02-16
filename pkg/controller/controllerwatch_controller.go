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

package controller

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	controllerv1alpha1 "github.com/cheeseandcereal/kubehoist/api/v1alpha1"
	"github.com/cheeseandcereal/kubehoist/pkg/helm"
	"github.com/go-logr/logr"
)

// ControllerWatchReconciler reconciles a ControllerWatch object
type ControllerWatchReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	HelmClient *helm.HelmClient
}

// +kubebuilder:rbac:groups=controller.kubehoist.io,resources=controllerwatches,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controller.kubehoist.io,resources=controllerwatches/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=controller.kubehoist.io,resources=controllerwatches/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ControllerWatch object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *ControllerWatchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var controllerWatchResource controllerv1alpha1.ControllerWatch
	if err := r.Get(ctx, req.NamespacedName, &controllerWatchResource); err != nil {
		log.Error(err, "Unable to fetch controller watch custom resource")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// during a requeue.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if controllerWatchResource.Status.ControllerInstallationStatus == controllerv1alpha1.ControllerInstallationStatusInstalled {
		// This controller has already been installed. Nothing to do
		// TODO: Add any sort of health checks or update logic
		return ctrl.Result{}, nil
	}

	if controllerWatchResource.Status.CRDsInstallationStatus != controllerv1alpha1.CRDInstallationStatusInstalled {
		err := r.installCRDs(ctx, &controllerWatchResource, log)
		return ctrl.Result{}, err
	}

	if controllerWatchResource.Status.ControllerInstallationStatus == controllerv1alpha1.ControllerInstallationStatusPending {
		// Trigger installation of helm chart if the status of this controller installation is pending
		err := r.installController(ctx, &controllerWatchResource, log)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ControllerWatchReconciler) installCRDs(ctx context.Context, controllerWatchResource *controllerv1alpha1.ControllerWatch, log logr.Logger) error {
	log.Info("Installing CRDs from chart", "chart", controllerWatchResource.Spec.HelmControllerSpec.Chart)
	helmInstallOpts, err := getHelmInstallOptions(controllerWatchResource, log)
	if err != nil {
		err = r.updateCRDInstallationStatus(ctx, controllerWatchResource, controllerv1alpha1.CRDInstallationStatusInvalidHelmChartValues)
		return err
	}
	installedCRDs, err := r.HelmClient.InstallChartCRDs(ctx, helmInstallOpts, r.Client)
	if err != nil {
		log.Error(err, "Failed to template helm chart")
		err = r.updateCRDInstallationStatus(ctx, controllerWatchResource, controllerv1alpha1.CRDInstallationStatusHelmChartFailed)
		return err
	}
	if len(installedCRDs) == 0 {
		log.Error(err, "No CRDs found in helm chart")
		err = r.updateCRDInstallationStatus(ctx, controllerWatchResource, controllerv1alpha1.CRDInstallationStatusNoCRDsFound)
		return err
	}
	log.Info("Successfully installed CRDs from helm chart", "crds", installedCRDs)
	// TODO: Register new CRDs to a new watcher somewhere
	installed := []controllerv1alpha1.GroupVersionKind{}
	for _, crd := range installedCRDs {
		installed = append(installed, controllerv1alpha1.GroupVersionKind{
			Group:   crd.Group,
			Version: crd.Version,
			Kind:    crd.Kind,
		})
	}
	controllerWatchResource.Status.InstalledCRDs = installed
	err = r.updateCRDInstallationStatus(ctx, controllerWatchResource, controllerv1alpha1.CRDInstallationStatusInstalled)
	return err
}

func (r *ControllerWatchReconciler) installController(ctx context.Context, controllerWatchResource *controllerv1alpha1.ControllerWatch, log logr.Logger) error {
	log.Info("Installing Chart", "chart", controllerWatchResource.Spec.HelmControllerSpec.Chart)
	helmInstallOpts, err := getHelmInstallOptions(controllerWatchResource, log)
	if err != nil {
		err = r.updateControllerInstallationStatus(ctx, controllerWatchResource, controllerv1alpha1.ControllerInstallationStatusInstallFailed)
		return err
	}
	err = r.HelmClient.InstallChart(ctx, helmInstallOpts)
	if err != nil {
		log.Error(err, "Failed to install helm chart")
		err = r.updateControllerInstallationStatus(ctx, controllerWatchResource, controllerv1alpha1.ControllerInstallationStatusInstallFailed)
		return err
	}
	log.Info("Successfully installed helm chart")
	err = r.updateControllerInstallationStatus(ctx, controllerWatchResource, controllerv1alpha1.ControllerInstallationStatusInstalled)
	return err
}

func (r *ControllerWatchReconciler) updateCRDInstallationStatus(ctx context.Context, controllerWatchResource *controllerv1alpha1.ControllerWatch, status controllerv1alpha1.CRDInstallationStatus) error {
	controllerWatchResource.Status.CRDsInstallationStatus = status
	return r.updateStatus(ctx, controllerWatchResource)
}

func (r *ControllerWatchReconciler) updateControllerInstallationStatus(ctx context.Context, controllerWatchResource *controllerv1alpha1.ControllerWatch, status controllerv1alpha1.ControllerInstallationStatus) error {
	controllerWatchResource.Status.ControllerInstallationStatus = status
	return r.updateStatus(ctx, controllerWatchResource)
}

func (r *ControllerWatchReconciler) updateStatus(ctx context.Context, controllerWatchResource *controllerv1alpha1.ControllerWatch) error {
	controllerWatchResource.Status.LastUpdated = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, controllerWatchResource); err != nil {
		log.FromContext(ctx).Error(err, "could not update ControllerWatch status")
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControllerWatchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&controllerv1alpha1.ControllerWatch{}).
		Named("controllerwatch").
		Complete(r)
}

func getHelmInstallOptions(controllerWatchResource *controllerv1alpha1.ControllerWatch, log logr.Logger) (helm.InstallOptions, error) {
	values := map[string]interface{}{}
	if controllerWatchResource.Spec.HelmControllerSpec.Values != "" {
		err := yaml.Unmarshal([]byte(controllerWatchResource.Spec.HelmControllerSpec.Values), &values)
		if err != nil {
			log.Error(err, "Failed to unmarshal values")
			return helm.InstallOptions{}, err
		}
	}
	createNamespace := false
	if controllerWatchResource.Spec.HelmControllerSpec.CreateNamespace != nil {
		createNamespace = *controllerWatchResource.Spec.HelmControllerSpec.CreateNamespace
	}
	return helm.InstallOptions{
		ChartName:       controllerWatchResource.Spec.HelmControllerSpec.Chart,
		Namespace:       controllerWatchResource.Spec.HelmControllerSpec.Namespace,
		ReleaseName:     controllerWatchResource.Spec.HelmControllerSpec.ReleaseName,
		Version:         controllerWatchResource.Spec.HelmControllerSpec.Version,
		Values:          values,
		CreateNamespace: createNamespace,
	}, nil
}
