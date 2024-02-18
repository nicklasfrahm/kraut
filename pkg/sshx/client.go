package sshx

import (
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
)

// NewClient creates a new SSH client with all available SSH keys
// in the user's home directory via the default port 22.
func NewClient(host string, options ...Option) (*ssh.Client, error) {
	cfg, err := NewOptions(options...)
	if err != nil {
		return nil, err
	}

	// Create an SSH client config with the specified authentication methods
	config := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            cfg.AuthMethods,
		HostKeyCallback: cfg.HostKeyCallback,
	}

	address := net.JoinHostPort(host, fmt.Sprint(cfg.Port))
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection: %s", err)
	}

	return client, nil
}

// ProbeSSHHostPublicKeyFingerprint returns the SHA256 fingerprint of the host.
// The returned string is in the format "SHA256:<hash>".
func ProbeSSHHostPublicKeyFingerprint(host string) (string, error) {
	var fingerprint string
	extractor := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		fingerprint = ssh.FingerprintSHA256(key)
		return nil
	}

	client, err := NewClient(host,
		WithDefaultCredentials(),
		WithHostKeyCallback(ssh.HostKeyCallback(extractor)),
	)
	if err != nil {
		return "", err
	}
	defer client.Close()

	return fingerprint, nil
}
