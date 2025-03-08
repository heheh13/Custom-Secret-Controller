/*
Copyright 2025 heheh.

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

	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/heheh13/Custom-Secret-Controller/api/v1alpha1"
	corev1alpha1 "github.com/heheh13/Custom-Secret-Controller/api/v1alpha1"
	"github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomSecretReconciler reconciles a CustomSecret object
type CustomSecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.heheh.org,resources=customsecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.heheh.org,resources=customsecrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.heheh.org,resources=customsecrets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CustomSecret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *CustomSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	cs := &v1alpha1.CustomSecret{}

	err := r.Get(ctx, req.NamespacedName, cs)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Infoln("Object customsecret ", req.NamespacedName, "might be deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	klog.Infoln("CustomSecret found", req.NamespacedName)

	if cs.Spec.SecretType != corev1.SecretTypeBasicAuth {
		klog.Infoln("Skipping CustomSecret is not of type BasicAuth", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	// Check if the secret exists
	secret := &corev1.Secret{}
	errSec := r.Client.Get(ctx, req.NamespacedName, secret)

	if errSec == nil { // Secret exists, check rotation time
		if !r.shouldRotateSecret(cs, req) {
			klog.Info("Secret rotation not needed yet", req.NamespacedName)
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}
	}

	newSecret, err := r.createNewSecret(cs, req)
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	if errors.IsNotFound(errSec) {
		if err := r.Client.Create(ctx, newSecret); err != nil {
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}
		klog.Infoln("Secret created", req.NamespacedName)
	} else {
		if err := r.Client.Update(ctx, newSecret); err != nil {
			return ctrl.Result{RequeueAfter: time.Minute}, err
		}
		klog.Infoln("Secret updated", req.NamespacedName)
	}

	// Update LastUpdatedTimestamp in CustomSecret
	r.setLastUpdatetime(cs, req)
	if err := r.Status().Update(ctx, cs); err != nil {
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	klog.Infoln("Reconciliation complete", req.NamespacedName)

	return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
}

func (r *CustomSecretReconciler) shouldRotateSecret(cs *v1alpha1.CustomSecret, req ctrl.Request) bool {
	now := time.Now()
	lastUpdated := r.getLastupdatedtimeStamp(cs, req)
	timeDifference := now.Sub(lastUpdated.Time)

	klog.Infoln(req.NamespacedName, "time since last update: ", timeDifference)
	return timeDifference >= cs.Spec.RotationTime
}

func (r *CustomSecretReconciler) createNewSecret(cs *v1alpha1.CustomSecret, req ctrl.Request) (*corev1.Secret, error) {
	pass, err := password.Generate(40, 10, 10, false, false)
	if err != nil {
		return nil, err
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cs.Name,
			Namespace: cs.Namespace,
		},
		Type: corev1.SecretTypeBasicAuth,
		Data: map[string][]byte{
			"username": []byte("admin"),
			"password": []byte(pass),
		},
	}, nil
}

func (r *CustomSecretReconciler) getLastupdatedtimeStamp(customSecret *v1alpha1.CustomSecret, req ctrl.Request) metav1.Time {

	for _, secret := range customSecret.Status.UpdatedSecrets {

		if secret.Name == req.Name && secret.Namespace == req.Namespace {
			return secret.UpdatedAt
		}
	}
	return metav1.Time{}
}

func (r *CustomSecretReconciler) setLastUpdatetime(customSecret *v1alpha1.CustomSecret, req ctrl.Request) {
	for index, secret := range customSecret.Status.UpdatedSecrets {
		if secret.Name == req.Name && secret.Namespace == req.Namespace {
			customSecret.Status.UpdatedSecrets[index].UpdatedAt = metav1.Now()
			return
		}
	}
	updateStatus := v1alpha1.SecretUpdateStatus{
		Name:      req.Name,
		Namespace: req.Namespace,
		UpdatedAt: metav1.Now(),
	}

	customSecret.Status.UpdatedSecrets = append(customSecret.Status.UpdatedSecrets, updateStatus)

}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.CustomSecret{}).
		Named("customsecret").
		Complete(r)
}
