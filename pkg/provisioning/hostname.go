package provisioning

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/nicklasfrahm/kraut/pkg/log"
	"github.com/nicklasfrahm/kraut/pkg/sshx"
)

// GetHostname returns the hostname of the remote host.
func GetHostname(ctx context.Context, host string) (string, error) {
	client, err := sshx.NewClient(host,
		// TODO: Allow specifying the host key callback.
		sshx.WithHostKeyCallback(ssh.InsecureIgnoreHostKey()),
		sshx.WithDefaultCredentials(),
	)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	hostnameBytes, err := session.Output("hostname")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(hostnameBytes)), nil
}

// SetHostname sets the hostname of the remote host.
func SetHostname(ctx context.Context, host, desired string) error {
	logger := log.NewSingletonLogger()

	current, err := GetHostname(ctx, host)
	if err != nil {
		return err
	}

	if current == desired {
		return nil
	}

	client, err := sshx.NewClient(host,
		sshx.WithHostKeyCallback(ssh.InsecureIgnoreHostKey()),
		sshx.WithDefaultCredentials(),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	hostnamectlSession, err := client.NewSession()
	if err != nil {
		return err
	}
	defer hostnamectlSession.Close()

	hostnamectlCommand := fmt.Sprintf("sudo hostnamectl hostname %s", desired)
	hostnamectlOutput, err := hostnamectlSession.CombinedOutput(hostnamectlCommand)
	if err != nil {
		logger.Error("failed to run hostnamectl: " + strings.TrimSpace(string(hostnamectlOutput)))
		return err
	}

	sedSession, err := client.NewSession()
	if err != nil {
		return err
	}
	defer sedSession.Close()

	sedCommand := fmt.Sprintf(`sudo sed -i 's/%s/%s/g' /etc/hosts`, current, desired)
	sedOutput, err := sedSession.CombinedOutput(sedCommand)
	if err != nil {
		logger.Error("failed to run sed: " + strings.TrimSpace(string(sedOutput)))
		return err
	}

	return nil
}
