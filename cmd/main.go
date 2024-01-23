/*
Copyright 2024.

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
	"fmt"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	firewallv1alpha1 "github.com/nicklasfrahm/kraut/api/firewall/v1alpha1"
	managementv1alpha1 "github.com/nicklasfrahm/kraut/api/management/v1alpha1"
	firewallcontroller "github.com/nicklasfrahm/kraut/internal/controller/firewall"
	managementcontroller "github.com/nicklasfrahm/kraut/internal/controller/management"
	//+kubebuilder:scaffold:imports
)

var (
	version = "dev"
	scheme  = runtime.NewScheme()
	// CLI flags.
	metricsAddr          string
	enableLeaderElection bool
	probeAddr            string
	logJSON              bool
	logLevel             string
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(managementv1alpha1.AddToScheme(scheme))
	utilruntime.Must(firewallv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-election", false, "Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&logJSON, "log-json", false, "Enabling this will ensure logs are in JSON format.")
	flag.StringVar(&logLevel, "log-level", "info", "Set the log level.")
	flag.Parse()

	setupLog := configureLogging()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: metricsAddr,
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: 9443,
		}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "60c7ccd3.kraut.nicklasfrahm.dev",
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
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&managementcontroller.HostReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Host")
		os.Exit(1)
	}
	if err = (&firewallcontroller.FirewallReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Firewall")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "failed to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "failed to set up ready check")
		os.Exit(1)
	}

	setupLog.WithValues("version", version).Info("Starting controller manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "failed to start controller manager")
		os.Exit(1)
	}
}

func configureLogging() logr.Logger {
	logger := logrus.New()
	if logJSON {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	// Use the log level from the CLI flag or default to info.
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(level)
	}

	// Create logr.Logger from logrus.Logger.
	mgrLog := logrusr.New(logger, logrusr.WithFormatter(func(i interface{}) interface{} {
		// Ensure that structs are printed as strings instead of JSON.
		return fmt.Sprintf("%+v", i)
	}))
	setupLog := mgrLog.WithName("setup")

	// Give user feedback if the log level is not set correctly.
	// Note that this needs to be done after the logger has been
	// configured.
	if err != nil {
		setupLog.Error(err, "defaulting to log level info")
	}

	// Pass the configured logger to the controller-runtime.
	ctrl.SetLogger(mgrLog)

	return setupLog
}
