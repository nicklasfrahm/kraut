package ssh

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/ini.v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nicklasfrahm/k3se/pkg/sshx"
	mgmtv1alpha1 "github.com/nicklasfrahm/kraut/api/management/v1alpha1"
	"github.com/nicklasfrahm/kraut/pkg/management/common"
)

// Client manages an appliance using SSH.
type Client struct {
	host *mgmtv1alpha1.Host
	ssh  *sshx.Client
	kube client.Client
	os   *mgmtv1alpha1.OSInfo
}

// NewClient creates a new client for a host.
func NewClient(host *mgmtv1alpha1.Host, options ...common.Option) (common.Client, error) {
	opts, err := common.GetDefaultOptions().Apply(options...)
	if err != nil {
		return nil, err
	}

	return &Client{
		host: host,
		kube: opts.KubernetesClient,
	}, nil
}

// Connect connects to the host.
func (c *Client) Connect() error {
	// Fetch credentials from secret.
	secretRef := types.NamespacedName{
		Namespace: c.host.Spec.SecretRef.Namespace,
		Name:      c.host.Spec.SecretRef.Name,
	}
	if secretRef.Namespace == "" {
		secretRef.Namespace = c.host.ObjectMeta.Namespace
	}

	secret := new(corev1.Secret)
	err := c.kube.Get(context.TODO(), secretRef, secret)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return fmt.Errorf("failed to read Secret: %s/%s", secretRef.Namespace, secretRef.Name)
		}
		return err
	}

	var proxy *sshx.Client
	if c.host.Spec.SSH.ProxyHost != "" {
		proxy, err = sshx.NewClient(&sshx.Config{
			Host:        c.host.Spec.SSH.ProxyHost,
			Port:        c.host.Spec.SSH.ProxyPort,
			Fingerprint: c.host.Spec.SSH.ProxyFingerprint,
			User:        c.host.Spec.SSH.ProxyUser,
			Key:         string(secret.Data["proxyKey"]),
			Passphrase:  string(secret.Data["proxyPassphrase"]),
			Password:    string(secret.Data["proxyPasswordInsecure"]),
		}, sshx.WithSTFPDisabled())
		if err != nil {
			return err
		}
	}

	c.ssh, err = sshx.NewClient(&sshx.Config{
		Host:        c.host.Spec.Host,
		Port:        c.host.Spec.Port,
		Fingerprint: c.host.Spec.SSH.Fingerprint,
		User:        c.host.Spec.SSH.User,
		Key:         string(secret.Data["key"]),
		Passphrase:  string(secret.Data["passphrase"]),
		Password:    string(secret.Data["passwordInsecure"]),
	}, sshx.WithProxy(proxy), sshx.WithSTFPDisabled())
	if err != nil {
		return err
	}

	// Probe the operating system.
	c.os, err = c.probeOS()
	if err != nil {
		return err
	}

	return nil
}

// Disconnect disconnects from the host.
func (c *Client) Disconnect() error {
	return c.ssh.Close()
}

// OS returns information about the operating system the host.
func (c *Client) OS() *mgmtv1alpha1.OSInfo {
	return c.os
}

// probeOS probes the operating system of the host.
func (c *Client) probeOS() (*mgmtv1alpha1.OSInfo, error) {
	osReleaseFile := "/etc/os-release"

	osSession, err := c.ssh.SSH.NewSession()
	if err != nil {
		return nil, err
	}
	defer osSession.Close()

	osReleaseRaw, err := osSession.Output(fmt.Sprintf("cat %s", osReleaseFile))
	if err != nil {
		return nil, err
	}
	osRelease, err := ini.Load(osReleaseRaw)
	if err != nil {
		return nil, err
	}

	// Parse the INI formatted file.
	section := osRelease.Section("")
	if section == nil {
		return nil, fmt.Errorf("failed to parse file: %s", osReleaseFile)
	}
	osNameKey := section.Key("NAME")
	if osNameKey == nil {
		return nil, fmt.Errorf("failed to parse OS name from file: %s", osReleaseFile)
	}
	osVersionKey := section.Key("VERSION_ID")
	if osVersionKey == nil {
		return nil, fmt.Errorf("failed to parse OS version from file: %s", osReleaseFile)
	}

	kernelSession, err := c.ssh.SSH.NewSession()
	if err != nil {
		return nil, err
	}
	defer kernelSession.Close()

	kernelVersionRaw, err := kernelSession.Output("uname -r")
	if err != nil {
		return nil, err
	}

	// Return "Unknown" if either key can't be parsed.
	return &mgmtv1alpha1.OSInfo{
		Name:          osNameKey.MustString("Unknown"),
		Version:       mgmtv1alpha1.OSVersion(osVersionKey.MustString("Unknown")),
		KernelVersion: strings.TrimSpace(string(kernelVersionRaw)),
	}, nil
}
