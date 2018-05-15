package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/tchaudhry91/cloudinventory/collector"
)

var china bool

// awsCmd represents the aws command
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Dump AWS inventory. Currently supports EC2/RDS",
	Run: func(cmd *cobra.Command, args []string) {
		path := cmd.Flag("path").Value.String()
		filter := cmd.Flag("filter").Value.String()
		if !validateAWSFilter(filter) {
			fmt.Printf("Invalid filter selected, please select a supported AWS service")
			return
		}

		col, err := collector.NewAWSCollector(china)
		if err != nil {
			fmt.Printf("Failed to create AWS collector: %v\n", err)
			return
		}

		// Create a map per service
		result := make(map[string]interface{})

		switch filter {
		case "ec2":
			err := collectEC2(col, result)
			if err != nil {
				return
			}
		case "rds":
			err := collectRDS(col, result)
			if err != nil {
				return
			}
		default:
			err := collectEC2(col, result)
			if err != nil {
				return
			}
			err = collectRDS(col, result)
			if err != nil {
				return
			}
		}
		fmt.Printf("Dumping to %s", path)
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

func validateAWSFilter(filter string) bool {
	validSlice := []string{
		"ec2",
		"rds",
		"",
	}
	for _, service := range validSlice {
		if filter == service {
			return true
		}
	}
	return false
}

func collectEC2(col collector.AWSCollector, result map[string]interface{}) error {
	instances, err := col.CollectEC2()
	if err != nil {
		fmt.Printf("Failed to gather EC2 Data: %v\n", err)
		return err
	}
	fmt.Printf("Gathered EC2 Instances across %d regions\n", len(instances))
	result["ec2"] = instances
	return nil
}

func collectRDS(col collector.AWSCollector, result map[string]interface{}) error {
	instances, err := col.CollectRDS()
	if err != nil {
		fmt.Printf("Failed to gather RDS Data: %v\n", err)
		return err
	}
	fmt.Printf("Gathered RDS Instances across %d regions\n", len(instances))
	result["rds"] = instances
	return nil
}

func init() {
	awsCmd.PersistentFlags().BoolVarP(&china, "include-china", "", false, "Include the China Partition")
	dumpCmd.AddCommand(awsCmd)
}
