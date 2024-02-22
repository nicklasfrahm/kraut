package provisioning

import (
	"context"
	"fmt"
	"net"

	"github.com/juju/juju/network/netplan"
	"golang.org/x/crypto/ssh"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/nicklasfrahm/kraut/pkg/log"
	"github.com/nicklasfrahm/kraut/pkg/sshx"
)

// ConfigureLoopbackInterface configures the loopback interface on the remote host.
func ConfigureLoopbackInterface(ctx context.Context, host string, subnet net.IPNet) error {
	client, err := sshx.NewClient(host,
		sshx.WithHostKeyCallback(ssh.InsecureIgnoreHostKey()),
		sshx.WithDefaultCredentials(),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	netplanYAML, _, err := client.RunCommand("netplan get")
	if err != nil {
		return err
	}

	var plan netplan.Netplan
	if err := yaml.UnmarshalStrict([]byte(netplanYAML), &plan); err != nil {
		return err
	}

	rebuilt, err := netplan.Marshal(&plan)
	if err != nil {
		return err
	}

	// TODO: Continue here.
	fmt.Printf("%+v\n", plan)
	fmt.Println(string(rebuilt))

	return nil
}

// ConfigureDHCPClient configures the DHCP client on the remote host.
func ConfigureDHCPClient(ctx context.Context, host string) error {
	logger := log.NewSingletonLogger()

	logger.Error("failed to find implementation")

	return nil
}

// ConfigureWANInterface configures the WAN interface on the remote host.
func ConfigureWANInterface(ctx context.Context, host string) error {
	logger := log.NewSingletonLogger()

	logger.Error("failed to find implementation")

	return nil
}
