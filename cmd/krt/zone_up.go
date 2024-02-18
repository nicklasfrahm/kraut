package main

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/nicklasfrahm/kraut/pkg/zone"
)

var (
	zoneUpCmdConfig = zone.Zone{
		Router: &zone.ZoneRouter{},
	}
	configFile string
)

var zoneUpCmd = &cobra.Command{
	Use:   "up <host>",
	Short: "Bootstrap a new availability zone",
	Long: `This command will bootstrap a new zone by connecting
to the specified IP and setting up a k3s cluster on
the host that will then set up the required services
for managing the lifecycle of the zone.

To manage a zone, the CLI needs credentials for the
DNS provider that is used to manage the DNS records
for the zone. These credentials can only be provided
via the environment variable DNS_PROVIDER_CREDENTIAL
and DNS_PROVIDER or via a ".env" file in the current
working directory.`,
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"host"},
	ValidArgs:  []string{"host"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// This should be safe because of the ExactArgs(1) constraint,
		// but we still need to check it to avoid panics.
		if len(args) != 1 {
			logger.Fatal("expected exactly one argument", zap.Strings("args", args))
		}
		host := args[0]

		if err := zone.Up(host, &zoneUpCmdConfig); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	zoneUpCmd.Flags().StringVarP(&zoneUpCmdConfig.Name, "name", "n", "", "name of the zone")
	zoneUpCmd.Flags().StringVarP(&zoneUpCmdConfig.Domain, "domain", "d", "", "domain that will contain the DNS records for the zone")
	zoneUpCmd.Flags().StringVarP(&zoneUpCmdConfig.Router.Hostname, "hostname", "H", "", "hostname of the router serving the zone")
	zoneUpCmd.Flags().IPVarP(&zoneUpCmdConfig.Router.ID, "router-id", "r", nil, "IPv4 address of the router serving the zone")
	zoneUpCmd.Flags().Uint32VarP(&zoneUpCmdConfig.Router.ASN, "asn", "a", 0, "autonomous system number of the zone")
	zoneUpCmd.Flags().StringVarP(&configFile, "config", "c", "", "path to the configuration file")
}
