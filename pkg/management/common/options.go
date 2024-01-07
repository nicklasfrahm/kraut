package common

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Options contains the configuration for an operation.
type Options struct {
	// KubernetesClient is a Kubernetes client that knows how to talk to the Kubernetes API.
	KubernetesClient client.Client
}

// Option applies a configuration option
// for the execution of an operation.
type Option func(options *Options) error

// Apply applies the option functions to the current set of options.
func (o *Options) Apply(options ...Option) (*Options, error) {
	for _, option := range options {
		if err := option(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

// GetDefaultOptions returns the default options
// for all operations of this library.
func GetDefaultOptions() *Options {
	return &Options{
		KubernetesClient: nil,
	}
}

// WithKubernetesClient allows to provide a Kubernetes
// client that knows how to talk to the Kubernetes API.
func WithKubernetesClient(kubernetesClient client.Client) Option {
	return func(options *Options) error {
		options.KubernetesClient = kubernetesClient
		return nil
	}
}
