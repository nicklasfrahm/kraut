package zone

import (
	"context"

	"github.com/nicklasfrahm/kraut/pkg/log"
	"github.com/nicklasfrahm/kraut/pkg/provisioning"
)

// Up creates or updates an availability zone.
func Up(host string, zone *Zone) error {
	logger := log.NewSingletonLogger()
	// TODO: Validate args:
	//       - IP (e.g. 212.112.144.171) [ephemeral, required]
	//       - hostname (e.g. alfa.nicklasfrahm.dev) [config, required]
	//       - zone name (e.g. aar1) [config, required]
	//       - router ID (e.g. 172.31.255.0) [config, required]
	//       - ASN (e.g. 65000) [config, required]
	//       - DNS provider credential [env only, required]
	//       - user account password for local recovery [env only, optional]
	logger.Info("Validating zone configuration")
	if err := zone.Validate(); err != nil {
		return err
	}

	logger.Info("Validating SSH connectivity")
	if err := PreflightCheckSSH(context.TODO(), host); err != nil {
		return err
	}

	logger.Info("Validating Kubernetes API connectivity")
	if err := PreflightCheckKubeAPI(context.TODO(), host); err != nil {
		return err
	}

	logger.Info("Setting hostname")
	if err := provisioning.SetHostname(context.TODO(), host, zone.Router.Hostname); err != nil {
		return err
	}

	// TODO: Perform minimal system configuration:
	//       - Reset user password (if provided)

	// We recommend a subnet of at least 64 IP addresses, so let's consider
	// the subnet `172.18.0.0/26` for this example. The loopback interface
	// will be assigned the first IP address of the subnet, because it will
	// be used as the gateway address for the subnet. So in this case, the
	// loopback interface will be assigned the IP address `172.18.0.1`.

	logger.Info("Configuring loopback interface")
	if err := provisioning.ConfigureLoopbackInterface(context.TODO(), host, zone.Router.GatewaySubnet); err != nil {
		return err
	}

	logger.Info("Configuring DHCP client")
	if err := provisioning.ConfigureDHCPClient(context.TODO(), host); err != nil {
		return err
	}

	logger.Info("Configuring WAN interface")
	if err := provisioning.ConfigureWANInterface(context.TODO(), host); err != nil {
		return err
	}

	// TODO: Ensure minimal interface configuration:
	//       - IPv4 on loopback
	//       - Identify WAN interface and name it WAN
	//       - DHCP on all interfaces (if not configured)
	// TODO: Install or upgrade k3s
	// TODO: Install or upgrade kraut

	logger.Info("Successfully reconciled availability zone")

	return nil
}
