package cmd

import (
	azurelib "github.com/adobe/cloudinventory/azurelib"
	"github.com/spf13/cobra"
)

// azureCmd: azure command. Gets added to rootCmd
var azureCmd = &cobra.Command{
	Use:   "azure",
	Short: "Extract Azure inventory. Currently supports VM's(arg: vm) and PostgresDB(arg: pg)",
	Long: `Use vm for virtual machine inventory from Azure and
		   use pg for extracting postgres data from Azure. 
		   If you dont provide args, both virtual machine and postgres data will be returned `,
	ValidArgs: []string{"vm", "pg"},
	Args:      matchAll(cobra.MinimumNArgs(0), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			for _, arg := range args {
				switch arg {
				case "vm":
					azurelib.ExtractVMInventory()
				case "pg":
					azurelib.ExtractPostgresInventory()
				default:
					azurelib.ExtractVMInventory()
					azurelib.ExtractPostgresInventory()
				}

			}
		} else {
			azurelib.ExtractVMInventory()
			azurelib.ExtractPostgresInventory()
		}
	},
}

func matchAll(allArgs ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, eachArg := range allArgs {
			if err := eachArg(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

func init() {
	rootCmd.AddCommand(azureCmd)
}
