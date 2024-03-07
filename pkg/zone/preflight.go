package zone

import (
	"context"
	"fmt"

	"github.com/nicklasfrahm/kraut/pkg/netutil"
)

// PreflightCheckSSH performs a preflight check for SSH connectivity.
// The port must be open for the check to pass, because we expect to
// have an SSH server running on the host.
// TODO: Can we turn this into a temporal activity, that can be used
// in the CLI?
func PreflightCheckSSH(ctx context.Context, host string) error {
	port := netutil.PortSSH
	if status := netutil.ProbeTCP(host, port); status != netutil.ProbeStatusOpen {
		return fmt.Errorf("failed to perform preflight check: port %d/tcp is %s", port, status)
	}

	return nil
}

// PreflightCheckKubeAPI performs a preflight check for the Kubernetes API.
// The port may not be `closed`, but it may be `filtered“. This is because we
// do not expect to have an API server running on the host yet. We do however
// expect to have a firewall in place that does not actively `drop` packets.
func PreflightCheckKubeAPI(ctx context.Context, host string) error {
	// We run the Kubernetes API server on an alternative port, because we want
	// to perform SNI routing with other Kubernetes clusters on the default port.
	port := 7443
	if status := netutil.ProbeTCP(host, port); !(status == netutil.ProbeStatusOpen || status == netutil.ProbeStatusFiltered) {
		return fmt.Errorf("failed to perform preflight check: port %d/tcp is %s", port, status)
	}

	return nil
}
