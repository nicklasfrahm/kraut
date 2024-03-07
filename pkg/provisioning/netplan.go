package provisioning

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"

	"github.com/juju/juju/network/netplan"
	"golang.org/x/crypto/ssh"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/nicklasfrahm/kraut/pkg/log"
	"github.com/nicklasfrahm/kraut/pkg/sshx"
)

// ConfigureLoopbackInterface configures the loopback interface on the remote host.
func ConfigureLoopbackInterface(ctx context.Context, host string, subnet net.IPNet) error {
	logger := log.NewSingletonLogger()

	client, err := sshx.NewClient(host,
		sshx.WithHostKeyCallback(ssh.InsecureIgnoreHostKey()),
		sshx.WithDefaultCredentials(),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	interfaces, err := GetInterfaces(ctx, client, host)
	if err != nil {
		return err
	}

	var loopbacks []net.Interface
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			loopbacks = append(loopbacks, iface)
		}
	}

	if len(loopbacks) > 1 {
		return fmt.Errorf("failed to update configuration: ambiguous loopback interfaces found")
	}

	logger.Info("Fetching current netplan configuration ...")
	netplanYAML, _, err := client.RunCommand("netplan get")
	if err != nil {
		return err
	}

	var config netplan.Netplan
	if err := yaml.UnmarshalStrict([]byte(netplanYAML), &config); err != nil {
		return err
	}

	// TODO: Continue here.

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

// GetInterfaces returns a map of network interfaces on the remote host.
func GetInterfaces(ctx context.Context, client *sshx.Client, host string) (map[string]net.Interface, error) {
	interfaceNamesStdout, interfaceNamesStderr, err := client.RunCommand("ls -1 /sys/class/net")
	if err != nil {
		if interfaceNamesStderr != "" {
			return nil, fmt.Errorf("%s: %s", err, interfaceNamesStderr)
		}
		return nil, err
	}

	interfaces := make(map[string]net.Interface)
	scanner := bufio.NewScanner(strings.NewReader(interfaceNamesStdout))
	for scanner.Scan() {
		name := scanner.Text()
		iface := net.Interface{Name: name}

		// Fetch the hardware address.
		hwAddressStdout, hwAddressStderr, err := client.RunCommand(fmt.Sprintf("cat /sys/class/net/%s/address", name))
		if err != nil {
			if hwAddressStderr != "" {
				return nil, fmt.Errorf("%s: %s", err, hwAddressStderr)
			}
			return nil, err
		}
		hwAddress, err := net.ParseMAC(strings.TrimSpace(hwAddressStdout))
		if err != nil {
			return nil, err
		}
		iface.HardwareAddr = hwAddress

		// Fetch the MTU.
		mtuStdout, mtuStderr, err := client.RunCommand(fmt.Sprintf("cat /sys/class/net/%s/mtu", name))
		if err != nil {
			if mtuStderr != "" {
				return nil, fmt.Errorf("%s: %s", err, mtuStderr)
			}
			return nil, err
		}
		mtu, err := strconv.Atoi(strings.TrimSpace(mtuStdout))
		if err != nil {
			return nil, err
		}
		iface.MTU = mtu

		// Fetch the flags.
		flagsStdout, flagsStderr, err := client.RunCommand(fmt.Sprintf("cat /sys/class/net/%s/flags", name))
		if err != nil {
			if flagsStderr != "" {
				return nil, fmt.Errorf("%s: %s", err, flagsStderr)
			}
			return nil, err
		}
		flags, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimSpace(flagsStdout), "0x"), 16, 32)
		if err != nil {
			return nil, err
		}
		iface.Flags = linkFlags(uint32(flags))

		// Fetch the interface index.
		indexStdout, indexStderr, err := client.RunCommand(fmt.Sprintf("cat /sys/class/net/%s/ifindex", name))
		if err != nil {
			if indexStderr != "" {
				return nil, fmt.Errorf("%s: %s", err, indexStderr)
			}
			return nil, err
		}
		index, err := strconv.Atoi(strings.TrimSpace(indexStdout))
		if err != nil {
			return nil, err
		}
		iface.Index = index

		interfaces[name] = iface
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read network interfaces: %s", err)
	}

	return interfaces, nil
}

// Wait for feedback on GitHub issue: https://github.com/golang/go/issues/65926
func linkFlags(rawFlags uint32) net.Flags {
	var f net.Flags
	if rawFlags&syscall.IFF_UP != 0 {
		f |= net.FlagUp
	}
	if rawFlags&syscall.IFF_RUNNING != 0 {
		f |= net.FlagRunning
	}
	if rawFlags&syscall.IFF_BROADCAST != 0 {
		f |= net.FlagBroadcast
	}
	if rawFlags&syscall.IFF_LOOPBACK != 0 {
		f |= net.FlagLoopback
	}
	if rawFlags&syscall.IFF_POINTOPOINT != 0 {
		f |= net.FlagPointToPoint
	}
	if rawFlags&syscall.IFF_MULTICAST != 0 {
		f |= net.FlagMulticast
	}
	return f
}

// GetInterfaceAddresses returns a list of network addresses on the specified interface.
func GetInterfaceAddresses(ctx context.Context, client *sshx.Client, host, iface string) ([]net.IPNet, error) {
	addressesStdout, addressesStderr, err := client.RunCommand(fmt.Sprintf("ip -j -o addr show dev %s", iface))
	if err != nil {
		if addressesStderr != "" {
			return nil, fmt.Errorf("%s: %s", err, addressesStderr)
		}
		return nil, err
	}

	var addresses []net.IPNet
	scanner := bufio.NewScanner(strings.NewReader(addressesStdout))
	for scanner.Scan() {
		var addr struct {
			AddrInfo []struct {
				Local     string `json:"local"`
				Prefixlen int    `json:"prefixlen"`
			} `json:"addr_info"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &addr); err != nil {
			return nil, err
		}
		for _, info := range addr.AddrInfo {
			_, ipNet, err := net.ParseCIDR(fmt.Sprintf("%s/%d", info.Local, info.Prefixlen))
			if err != nil {
				return nil, err
			}
			addresses = append(addresses, *ipNet)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read network addresses: %s", err)
	}

	return addresses, nil
}
