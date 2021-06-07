/*
Copyright 2021.

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

package main

import (
	"context"
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

var (
	scheme     = runtime.NewScheme()
	setupLog   = ctrl.Log.WithName("setup")
	podLog     = ctrl.Log.WithName("podlog")
	annotation = Getenv("WATCH_ANNOTATION", "timestamp")
	namespaces = Getenv("WATCH_NAMESPACES", "")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	// Init operator
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "3b32f4b0.podlog.lexmill99.net",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Create k8s client
	client, err := kube_client.New(ctrl.GetConfigOrDie(), kube_client.Options{})
	if err != nil {
		podLog.Error(err, "unable to create client")
		os.Exit(1)
	}

	// Reconciler
	r := &PodReconciler{
		client,
	}

	// Create controller
	c, err := controller.New("pod-watcher", mgr, controller.Options{Reconciler: r})
	if err != nil {
		podLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	// Watch pods,
	err = c.Watch(&source.Kind{Type: &v1.Pod{}}, &handler.EnqueueRequestForObject{}, predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			pod := event.Object
			// Filter namespace , the namespace can be multiple because of it
			// we are splitting namespaces by `,` and searching pod namespace on it
			if namespaces != "" {
				if !Contains(strings.Split(namespaces, ","), pod.GetNamespace()) {
					return false
				}
			} 

			// Filter annotation
			if annotation != "" {
				// Split annotation to key value pair
				// Check if pod annatation has the annatation key value
				anKV := strings.Split(annotation, "=")
				anno := pod.GetAnnotations()
				val, ok := anno[anKV[0]]
				if  !ok {
					return false
				}
				if val != anKV[1] {
					return false
				}
			}

			// If pod created before 60 seconds don't process
			if time.Now().Sub(pod.GetCreationTimestamp().Time).Seconds() > 60 {
				return false
			}

			podLog.Info("Pod created", "name", pod.GetName())
			return true
		},
		// Don't handle other operations
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return false
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return false
		},
	})
	if err != nil {
		podLog.Error(err, "unable to create controller")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

type PodReconciler struct {
	kube_client.Client
}

func (r *PodReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	//// Prevent conflict for pod status
	time.Sleep(2)

	// Get new pod data
	pod := &v1.Pod{}
	err := r.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		podLog.Error(err, "Get error")
		return reconcile.Result{RequeueAfter: time.Second * 10}, err
	}

	// Add annatation
	timePod := strconv.FormatInt(time.Now().Unix(), 10)
	pod.SetAnnotations(map[string]string{"timestamp": timePod})

	// Update pod resource
	err = r.Update(ctx, pod, &kube_client.UpdateOptions{})
	if err != nil {
		podLog.Info("pod update error retrying")
		return reconcile.Result{RequeueAfter: time.Second * 10}, err
	}
	podLog.Info("Pod updated", "name", pod.Name, "time", timePod)

	return reconcile.Result{}, nil
}

func Getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
