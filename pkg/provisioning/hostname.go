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

	hostnameStdout, _, err := client.RunCommand("hostnamectl hostname")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(hostnameStdout), nil
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

	hostnamectlCommand := fmt.Sprintf("sudo hostnamectl hostname %s", desired)
	_, hostnamectlStderr, err := client.RunCommand(hostnamectlCommand)
	if err != nil {
		if hostnamectlStderr == "" {
			return err
		}
		logger.Error("failed to run hostnamectl: " + strings.TrimSpace(hostnamectlStderr))
		return err
	}

	sedCommand := fmt.Sprintf(`sudo sed -i 's/%s/%s/g' /etc/hosts`, current, desired)
	_, sedStderr, err := client.RunCommand(sedCommand)
	if err != nil {
		if sedStderr == "" {
			return err
		}
		logger.Error("failed to run sed: " + strings.TrimSpace(sedStderr))
		return err
	}

	return nil
}
