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

package firewall

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	fwv1alpha1 "github.com/nicklasfrahm/kraut/api/firewall/v1alpha1"
	mgmtv1alpha1 "github.com/nicklasfrahm/kraut/api/management/v1alpha1"
)

const (
	controllerName = "firewall-controller"
)

// FirewallReconciler reconciles a Firewall object
type FirewallReconciler struct {
	client.Client
	recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=firewall.kraut.nicklasfrahm.dev,resources=firewalls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=firewall.kraut.nicklasfrahm.dev,resources=firewalls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=firewall.kraut.nicklasfrahm.dev,resources=firewalls/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Firewall object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *FirewallReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	firewall := new(fwv1alpha1.Firewall)
	if err := r.Get(ctx, req.NamespacedName, firewall); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	hostList := new(mgmtv1alpha1.HostList)
	if err := r.List(ctx, hostList); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO: Should we use pointers here to avoid inflating the memory usage?
	hosts := make([]mgmtv1alpha1.Host, 0)
	for _, host := range hostList.Items {
		err, isMatch := firewall.Spec.HostSelector.MatchMetadata.Matches(&host.ObjectMeta)
		if err != nil {
			r.recorder.Event(firewall, "Warning", "InvalidHostSelector", err.Error())
			// Deliberately fail the firewall reconciliation
			// to avoid a partial and insecure firewall setup.
			return ctrl.Result{}, err
		}

		if isMatch {
			if err := r.checkHostCompatibility(ctx, &host); err != nil {
				r.recorder.Event(firewall, "Warning", "HostIncompatible", err.Error())
				return ctrl.Result{}, err
			}
			hosts = append(hosts, host)
		}
	}

	if firewall.Status.HostCount != len(hosts) {
		firewall.Status.HostCount = len(hosts)
		if err := r.Status().Update(ctx, firewall); err != nil {
			return ctrl.Result{}, err
		}
	}

	// TODO(user): Implement reconciliation logic.
	// 2. For each host:
	// 4. Ensure that there is a firewall rule for the management protocol (to prevent lockout).
	// 5. Convert the intents into ruleset.
	// 6. Filter ruleset to only include relevant for firewall configuration for the host.
	// 7. Compare the ruleset with the current ruleset on the host.
	// 8. If there is a difference, apply the ruleset to the host.
	// 9. Check if the firewall service is enabled and running on the host.

	// TODO: How do we handle lifecycle of the host if is no longer selected?

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FirewallReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor(controllerName)

	return ctrl.NewControllerManagedBy(mgr).
		For(&fwv1alpha1.Firewall{}).
		Complete(r)
}

// verifyHostCompatibility checks if the host is compatible with the currently supported drivers.
// TODO: Implement abstract driver interface as part of "pkg/libintent".
func (r *FirewallReconciler) checkHostCompatibility(ctx context.Context, host *mgmtv1alpha1.Host) error {
	hostReference := fmt.Sprintf("%s/%s", host.ObjectMeta.Namespace, host.ObjectMeta.Name)

	// Ubuntu started supporting "nftables" in the 21.04 release.
	// Reference: https://lwn.net/Articles/867185/
	// TODO: Improve this, maybe using semver comparison, because this is cursed.
	if host.Status.OS.Name == mgmtv1alpha1.OSUbuntu {
		errUbuntu2104Required := fmt.Errorf("failed to detect supported OS version: Ubuntu 21.04 or later is required: %s", hostReference)

		if host.Status.OS.Version.Major() < 21 {
			return errUbuntu2104Required
		}

		if host.Status.OS.Version.Major() == 21 && host.Status.OS.Version.Minor() < 4 {
			return errUbuntu2104Required
		}

		return nil
	}

	return fmt.Errorf("failed to detect supported OS: %s", hostReference)
}
