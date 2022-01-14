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
	"net"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/zhangkunpeng/kubepxe/api/v1"
	"github.com/zhangkunpeng/kubepxe/dhcp"
)

const (
	v4               = 4
	v6               = 6
	EventTypeWarning = "Warning"
	EventTypeNormal  = "Normal"
)

// DynamicHostConfigurationReconciler reconciles a DynamicHostConfiguration object
type DynamicHostConfigurationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=pxe.kumple.com,resources=dynamichostconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pxe.kumple.com,resources=dynamichostconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pxe.kumple.com,resources=dynamichostconfigurations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DynamicHostConfiguration object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *DynamicHostConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx, "dhcp", req.String())

	obj := &v1.DynamicHostConfiguration{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if errors.IsNotFound(err) {
			dhcp.Close(req.String())
			err = nil
		}
		return ctrl.Result{}, err
	}
	patch := client.MergeFrom(obj.DeepCopy())
	var update = func() {
		if err := r.Status().Patch(ctx, obj, patch); err != nil {
			logger.Error(err, "Patch DynamicHostConfiguration failed")
		}
	}

	conf, err := load(obj.Spec)
	if err != nil {
		r.Recorder.Event(obj, EventTypeWarning, "LoadDHCPConfig", err.Error())
		return ctrl.Result{}, nil
	}

	svc, err := dhcp.Start(req.String(), conf)
	if err != nil {
		r.Recorder.Event(obj, EventTypeWarning, "StartDHCPServer", err.Error())
		obj.Status.State = "Failed"
		update()
		return ctrl.Result{}, nil
	}
	if svc != nil {
		go r.wait(req.NamespacedName, svc)
	}

	obj.Status.State = "Running"
	update()
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DynamicHostConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("DHCP")
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.DynamicHostConfiguration{}).
		Complete(r)
}

func (r *DynamicHostConfigurationReconciler) wait(nn types.NamespacedName, svc *dhcp.Server) {
	err := svc.Wait()
	obj := &v1.DynamicHostConfiguration{}
	ctx := context.Background()
	logger := log.FromContext(ctx, "dhcp", nn.String())
	if err := r.Get(ctx, nn, obj); err != nil {
		logger.Error(err, "DynamicHostConfiguration get failed")
		return
	}
	if err == nil {
		r.Recorder.Event(obj, EventTypeNormal, "DHCPServerClosed", "dhcp server closed")
		return
	}
	r.Recorder.Eventf(obj, EventTypeWarning, "DHCPServerStoped", "dhcp server stopped, err: %v", err)
	patch := client.MergeFrom(obj.DeepCopy())
	obj.Status.State = "Stoped"
	if err := r.Status().Patch(ctx, obj, patch); err != nil {
		logger.Error(err, "Patch DynamicHostConfiguration failed")
	}
}

func load(spec v1.DynamicHostConfigurationSpec) (*dhcp.Config, error) {
	sc, err := loadConfig(spec)
	if err != nil {
		return nil, err
	}
	conf := dhcp.NewConfig(spec.ProtocolVersion, sc)
	return conf, nil
}

func loadConfig(spec v1.DynamicHostConfigurationSpec) (*dhcp.ServerConfig, error) {
	listeners, err := parseListen(spec.ProtocolVersion, spec.Listen)
	if err != nil {
		return nil, err
	}
	plugins := make([]dhcp.PluginConfig, 0)

	sc := dhcp.NewServerConfig(listeners, plugins)
	return sc, nil
}

func parseListen(ver int, bind []v1.Bind) ([]net.UDPAddr, error) {
	listeners := []net.UDPAddr{}
	for _, b := range bind {
		ip := net.ParseIP(b.Address)
		if b.Address == "" {
			switch ver {
			case v4:
				ip = net.IPv4zero
			case v6:
				ip = net.IPv6unspecified
			}
		}

		if ip == nil {
			return nil, fmt.Errorf("dhcpv%d: invalid IP address in `listen` directive: %s", ver, b.Address)
		}
		if ip4 := ip.To4(); (ver == v6 && ip4 != nil) || (ver == v4 && ip4 == nil) {
			return nil, fmt.Errorf("dhcpv%d: not a valid IPv%d address in `listen` directive: '%s'", ver, ver, b.Address)
		}
		var port = b.Port
		if port == 0 {
			switch ver {
			case v4:
				port = dhcp.DefaultServerPortV4
			case v6:
				port = dhcp.DefaultServerPortV6
			}
		}
		if port > 65535 {
			return nil, fmt.Errorf("dhcpv%d: invalid `listen` port '%d'", ver, port)
		}
		l := net.UDPAddr{
			IP:   ip,
			Port: port,
			Zone: b.Interface,
		}
		listeners = append(listeners, l)
	}
	return listeners, nil
}
