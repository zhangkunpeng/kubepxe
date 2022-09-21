/*
Copyright 2022 Zhang Kunpeng.

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
	"fmt"
	pxev1 "github.com/zhangkunpeng/kubepxe/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// IPAddressReconciler reconciles a IPAddress object
type IPAddressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pxe.kumple.com,resources=ipaddresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pxe.kumple.com,resources=ipaddresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pxe.kumple.com,resources=ipaddresses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the IPAddress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *IPAddressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	obj := &pxev1.IPAddress{}
	err := r.Get(ctx, req.NamespacedName, obj)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}
	l.Info("current", "host", obj.Spec.Host)
	// TODO(user): your logic here
	f := func(h string) {
		l.Info("update:", "host", h)
		obj := &pxev1.IPAddress{}
		r.Get(ctx, req.NamespacedName, obj)
		obj.Spec.Host = h
		err := r.Update(ctx, obj)
		//debug.PrintStack()
		if err != nil {
			l.Error(err, h)
			return
		}
		obj1 := &pxev1.IPAddress{}
		r.Get(ctx, req.NamespacedName, obj1)
		l.Info("new", "host", obj.Spec.Host)
	}
	if obj.Spec.Host == "" {
		for i := 0; i < 4; i++ {
			f(fmt.Sprintf("host%d", i))
			time.Sleep(time.Second)
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IPAddressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pxev1.IPAddress{}).
		Complete(r)
}
