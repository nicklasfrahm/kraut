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

	// TODO: Ensure minimal interface configuration:
	//       - IPv4 on loopback
	//       - Identify WAN interface and name it WAN
	//       - DHCP on all interfaces (if not configured)
	// TODO: Install or upgrade k3s
	// TODO: Install or upgrade kraut

	logger.Info("Successfully reconciled availability zone")

	return nil
}
