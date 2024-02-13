package zone

import (
	"fmt"
	"net"

	"github.com/nicklasfrahm/kraut/pkg/netutil"
	"github.com/nicklasfrahm/kraut/pkg/sshx"
)

// Up creates or updates an availability zone.
func Up(host string, zone *Zone) error {
	// TODO: Validate args:
	//       - IP (e.g. 212.112.144.171) [ephemeral, required]
	//       - hostname (e.g. alfa.nicklasfrahm.dev) [config, required]
	//       - zone name (e.g. aar1) [config, required]
	//       - router ID (e.g. 172.31.255.0) [config, required]
	//       - ASN (e.g. 65000) [config, required]
	//       - DNS provider credential [env only, required]
	//       - user account password for local recovery [env only, optional]
	if err := zone.Validate(); err != nil {
		return err
	}

	if status := netutil.ProbeTCP(net.JoinHostPort(host, fmt.Sprint(netutil.PortSSH))); status != netutil.ProbeStatusOpen {
		return fmt.Errorf("failed to perform preflight check: port %d/tcp is %s", netutil.PortSSH, status)
	}

	{
		// TODO: Continue here.
		client, err := sshx.NewClient(host)
		if err != nil {
			return err
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			return err
		}
		defer session.Close()

		hostnameBytes, err := session.Output("hostname")
		if err != nil {
			return err
		}
		fmt.Printf("hostname detected: %s\n", string(hostnameBytes))
	}

	// TODO: Run preflight checks:
	//       - Open ports: TCP:22,80,443,6443,7443
	//       - Open ports: UDP:5800-5810
	// TODO: Continue here. Improve preflight checks:
	// 			 How do we differentiate between closed and filtered ports?
	tcpPorts := []int{22, 80, 443, 6443, 7443}
	for _, port := range tcpPorts {
		fmt.Printf("preflight check: port %4d: %s\n", port, netutil.ProbeTCP(host, port))
	}

	// TODO: Perform minimal system configuration:
	//       - Set hostname
	//       - Reset user password (if provided)
	// TODO: Ensure minimal interface configuration:
	//       - IPv4 on loopback
	//       - Identify WAN interface and name it WAN
	//       - DHCP on all interfaces (if not configured)
	// TODO: Install or upgrade k3s
	// TODO: Install or upgrade kraut

	fingerprint, err := sshx.GetSSHHostPublicKeyFingerprint(host)
	if err != nil {
		return err
	}

	fmt.Printf("fingerprint detected: %s: %s\n", host, fingerprint)

	return nil
}
