/*
Copyright 2023.

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
	"flag"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/AppsFlyer/local-pvc-releaser/internal/controller"
	"github.com/AppsFlyer/local-pvc-releaser/internal/exporters"
	"github.com/AppsFlyer/local-pvc-releaser/internal/initializers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var dryrun bool
	var devLogging bool
	var pvcSelector bool
	var pvcAnoCustomKey string
	var pvcAnoCustomValue string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&dryrun, "dry-run", false, "Enable controller in dry-run mode.")
	flag.BoolVar(&devLogging, "dev-logging", false, "Enable controller logger in dev format with stack tracing.")
	flag.BoolVar(&pvcSelector, "enable-pvc-selector", false, "Manage only PVC objects marked with custom annotation.")
	flag.StringVar(&pvcAnoCustomKey, "pvc-annotation-custom-key", "appsflyer.com/local-pvc-releaser", "PVC Annotations filter key.")
	flag.StringVar(&pvcAnoCustomValue, "pvc-annotation-custom-value", "enabled", "PVC Annotations filter value.")
	flag.Parse()

	logger, err := initializers.NewLogger(devLogging, dryrun)
	if err != nil {
		setupLog.Error(err, "failed to initialize logger")
		os.Exit(1)
	}
	if dryrun {
		logger.Info("controller started in dry-run mode")
	}

	ctrl.SetLogger(*logger)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "ab49af34.appsflyer.com",
		DryRunClient:           dryrun,
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Init Prometheus exporter and register it
	logger.Info("registering new collector metrics")
	collector := exporters.NewCollector()
	metrics.Registry.MustRegister(collector)

	if err = (&controller.PVCReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		Recorder:          mgr.GetEventRecorderFor("local-pvc-releaser"),
		DryRun:            dryrun,
		PvcSelector:       pvcSelector,
		PvcAnoCustomKey:   pvcAnoCustomKey,
		PvcAnoCustomValue: pvcAnoCustomValue,
		Logger:            logger,
		Collector:         collector,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PersistentVolumeClaim")
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
