package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

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
		path := cmd.Flag("path").Value.String()
		result := make(map[string]interface{})

		if len(args) > 0 {
			for _, arg := range args {
				switch arg {
				case "vm":
					vmlist, err := azurelib.ExtractVMInventory()
					if err != nil {
						fmt.Printf("Cannot obtain VM information - %v \n", err)
					} else {
						result["vm"] = vmlist
					}
				case "pg":
					pglist, err := azurelib.ExtractPostgresInventory()
					if err != nil {
						fmt.Printf("Cannot obtain Postgres information - %v \n", err)
					} else {
						result["pg"] = pglist
					}
				default:
					vmlist, err := azurelib.ExtractVMInventory()
					if err != nil {
						fmt.Printf("Cannot obtain VM information - %v \n", err)
					} else {
						result["vm"] = vmlist
					}
					pglist, err := azurelib.ExtractPostgresInventory()
					if err != nil {
						fmt.Printf("Cannot obtain Postgres information - %v \n", err)
					} else {
						result["pg"] = pglist
					}
				}

			}
		} else {
			vmlist, err := azurelib.ExtractVMInventory()
			if err != nil {
				fmt.Printf("Cannot obtain VM information - %v \n", err)
			} else {
				result["vm"] = vmlist
			}
			pglist, err := azurelib.ExtractPostgresInventory()
			if err != nil {
				fmt.Printf("Cannot obtain Postgres information - %v \n", err)
			} else {
				result["pg"] = pglist
			}
		}
		fmt.Printf("Writing data to %s \n", path)
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			fmt.Printf("Error Marshalling JSON: %v\n", err)
		}
		err = ioutil.WriteFile(path, jsonBytes, 0644)
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
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
	azureCmd.PersistentFlags().StringP("path", "p", "azurecloudinventory.json", "file path to dump the inventory in")
	rootCmd.AddCommand(azureCmd)
}
