package cmd

import (
	"github.com/spf13/cobra"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dumps the inventory for the given options",
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	dumpCmd.PersistentFlags().StringP("filter", "f", "", "limit dump to a particular cloud service, e.g ec2/rds")
	dumpCmd.PersistentFlags().StringP("path", "p", "cloudinventory.json", "file path to dump the inventory in")

}
