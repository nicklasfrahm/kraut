package common

import (
	mgmtv1alpha1 "github.com/nicklasfrahm/kraut/api/management/v1alpha1"
)

// ClientFactory is a function that creates a new client.
type ClientFactory func(*mgmtv1alpha1.Host, ...Option) (Client, error)

// Client is the interface for a client.
type Client interface {
	// Connect connects to the host.
	Connect() error
	// Disconnect disconnects from the host.
	Disconnect() error
	// OS returns information about the operating system the host.
	OS() *mgmtv1alpha1.OSInfo
}
