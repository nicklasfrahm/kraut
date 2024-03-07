package sshx

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Option is the interface of a functional option for the SSH client.
type Option func(*Config) error

// Config defines the configuration of the SSH client.
type Config struct {
	// AuthMethods is a list of authentication methods for the SSH connection.
	AuthMethods []ssh.AuthMethod
	// HostKeyCallback is a callback function that is used to verify the host key.
	HostKeyCallback ssh.HostKeyCallback
	// User is the username for the SSH connection.
	User string
	// Port is the port for the SSH connection.
	Port int
}

// NewOptions applies the options to the SSH client.
func NewOptions(options ...Option) (*Config, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	opts := &Config{
		AuthMethods: make([]ssh.AuthMethod, 0),
		User:        u.Username,
		Port:        22,
	}

	for _, option := range options {
		if err := option(opts); err != nil {
			return nil, err
		}
	}

	return opts, nil
}

// WithUserSSHKeys adds the user's SSH keys to the SSH client.
// If `requireKeys` is `true`, the function will return an
// error if no keys are found.
func WithUserSSHKeys(requireKeys bool) Option {
	return func(config *Config) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		if config.AuthMethods == nil {
			config.AuthMethods = make([]ssh.AuthMethod, 0)
		}

		sshDir := path.Join(home, ".ssh")
		files, err := os.ReadDir(sshDir)
		if err != nil {
			return err
		}

		signers := make([]ssh.Signer, 0)
		for _, entry := range files {
			// Avoid directories and public key files.
			if entry.IsDir() || strings.HasSuffix(entry.Name(), ".pub") {
				continue
			}

			privateKeyFile := filepath.Join(sshDir, entry.Name())
			privateKey, err := os.ReadFile(privateKeyFile)
			if err != nil {
				return err
			}

			signer, err := ssh.ParsePrivateKey(privateKey)
			if err != nil {
				// Ignore files that are not private keys.
				continue
			}

			signers = append(signers, signer)
		}

		if requireKeys && len(signers) == 0 {
			return fmt.Errorf("failed to find any private keys: %s", sshDir)
		}

		config.AuthMethods = append(config.AuthMethods, ssh.PublicKeys(signers...))

		return nil
	}
}

// WithHostKeyCallback adds a host key callback to the SSH client.
func WithHostKeyCallback(callback ssh.HostKeyCallback) Option {
	return func(config *Config) error {
		config.HostKeyCallback = callback
		return nil
	}
}

// WithUser sets the username for the SSH connection.
// The default username is the current user running
// the program.
func WithUser(user string) Option {
	return func(config *Config) error {
		config.User = user
		return nil
	}
}

// WithDefaultCredentials tries to automatically configure the SSH client.
func WithDefaultCredentials() Option {
	return func(config *Config) error {
		if err := WithUserSSHKeys(false)(config); err != nil {
			return err
		}

		// TODO: Allow authentication via OpenPubKey / OIDC.

		return nil
	}
}
