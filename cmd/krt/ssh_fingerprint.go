package main

import (
	"fmt"
	"os"
	"sync"
	"text/tabwriter"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"

	"github.com/nicklasfrahm/kraut/pkg/sshx"
)

var sshFingerprintCmd = &cobra.Command{
	Use:   "fingerprint <host> [...hosts]",
	Short: "Fetches the SSH fingerprint of a host",
	Long: `This command will fetch the SSH fingerprint of
the specified host and print it to the console.

The format is a SHA256 fingerprint of the host's
public key, similar to "SHA256:<hash>".`,
	Args:       cobra.MinimumNArgs(1),
	ArgAliases: []string{"host"},
	ValidArgs:  []string{"host"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// This should be safe because of the MinimumNArgs(1) constraint,
		// but we still need to check it to avoid panics.
		if len(args) < 1 {
			logger.Fatal("expected at lease 1 argument")
		}
		hosts := args

		// TODO: Move this into a function called
		// `ProbeSSHHostPublicKeyFingerprints` in
		// the sshx pacakge.
		var wg sync.WaitGroup
		fingerprints := make([]string, len(hosts))
		for i, host := range hosts {
			wg.Add(1)
			go func(i int, host string) {
				defer wg.Done()

				fingerprint, err := sshx.ProbeSSHHostPublicKeyFingerprint(host)
				if err != nil {
					fingerprint = err.Error()
				}

				fingerprints[i] = fingerprint
			}(i, host)
		}
		wg.Wait()

		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(tw, "HOST\tFINGERPRINT")
		for i, host := range hosts {
			fingerprint := fingerprints[i]
			fmt.Fprintf(tw, "%s\t%s\n", host, fingerprint)
		}
		tw.Flush()

		return nil
	},
}
