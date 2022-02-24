/*
Copyright 2022.

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
	dapps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	obmv1 "obm.datacommand/obmReplica/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ObmReplicaReconciler reconciles a ObmReplica object
type ObmReplicaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=obm.obm.datacommand,resources=obmreplicas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=obm.obm.datacommand,resources=obmreplicas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=obm.obm.datacommand,resources=obmreplicas/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ObmReplica object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ObmReplicaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	var obmReplica obmv1.ObmReplicaList
	if err := r.List(ctx, &obmReplica, client.InNamespace(req.Namespace)); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if len(obmReplica.Items) == 0 {
		return ctrl.Result{}, nil
	}

	var targetNamespace string
	for _, i := range obmReplica.Items {
		targetNamespace = i.Spec.TargetNamespace.Name
	}

	var deployments dapps.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(req.Namespace)); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	for _, i := range deployments.Items {
		i.Namespace = targetNamespace
		i.ResourceVersion = ""
		log.Log.Info("create deployment", targetNamespace, i.Name)
		if err := r.Create(ctx, &i); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ObmReplicaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&obmv1.ObmReplica{}).
		Owns(&dapps.Deployment{}).
		Watches(
			&source.Kind{Type: &dapps.Deployment{}},
			&handler.EnqueueRequestForObject{},
		).
		Complete(r)
}
