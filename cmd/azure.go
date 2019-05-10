package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	azurelib "github.com/adobe/cloudinventory/azurelib"
	"github.com/spf13/cobra"
)

// azureCmd: azure sub command. Gets added to dumpCmd
var azureCmd = &cobra.Command{
	Use:   "azure",
	Short: "Dump Azure inventory. Currently supports VM's(arg: vm) and PostgresDB(arg: pg)",
	Long:  `Use vm as an arg for virtual machine inventory from Azure and
	     use pg as an arg for extracting postgres data from Azure. 
	     If you dont provide args to --filter, both virtual machine and postgres data will be returned `,
	Run: func(cmd *cobra.Command, args []string) {
		path := cmd.Flag("path").Value.String()
		filter := cmd.Flag("filter").Value.String()
		result := make(map[string]interface{})

		if !validateAzureFilter(filter) {
			fmt.Printf("Invalid filter selected, please select a supported AWS service")
			return
		}

		switch filter {
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

func validateAzureFilter(filter string) bool {
	validSlice := []string{
		"vm",
		"pg",
		"",
	}
	for _, service := range validSlice {
		if filter == service {
			return true
		}
	}
	return false
}


func init() {
	dumpCmd.AddCommand(azureCmd)
}
