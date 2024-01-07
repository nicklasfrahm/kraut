package management

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mgmtv1alpha1 "github.com/nicklasfrahm/kraut/api/management/v1alpha1"
	"github.com/nicklasfrahm/kraut/pkg/management/common"
	"github.com/nicklasfrahm/kraut/pkg/management/ssh"
)

var clientFactories = map[mgmtv1alpha1.Protocol]common.ClientFactory{
	mgmtv1alpha1.ProtocolSSH: ssh.NewClient,
	// TODO: Add support for SNMP.
}

// NewClient returns a new client for the given host. Client will also implicitly
// connect to the appliance without an explicit call to Connect().
func NewClient(hostRef types.NamespacedName, options ...common.Option) (common.Client, error) {
	opts, err := common.GetDefaultOptions().Apply(options...)
	if err != nil {
		return nil, err
	}

	if opts.KubernetesClient == nil {
		return nil, fmt.Errorf("missing required option: WithKubernetesClient()")
	}

	host := new(mgmtv1alpha1.Host)
	err = opts.KubernetesClient.Get(context.TODO(), hostRef, host)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return nil, fmt.Errorf("failed to read Host: %s/%s", hostRef.Namespace, hostRef.Name)
		}
		return nil, err
	}

	newClient := clientFactories[host.Spec.Protocol]
	if newClient == nil {
		return nil, fmt.Errorf("unknown protocol: %s", host.Spec.Protocol)
	}

	mgmt, err := newClient(host, common.WithKubernetesClient(opts.KubernetesClient))
	if err != nil {
		return nil, err
	}

	if err := mgmt.Connect(); err != nil {
		return nil, err
	}

	return mgmt, nil
}
