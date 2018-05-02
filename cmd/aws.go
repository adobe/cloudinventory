// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/tchaudhry91/cloudinventory/collector"
)

// awsCmd represents the aws command
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Dump AWS inventory. Currently supports EC2/RDS",
	Run: func(cmd *cobra.Command, args []string) {
		path := cmd.Flag("path").Value.String()
		filter := cmd.Flag("filter").Value.String()
		if !validateAWSFilter(filter) {
			fmt.Errorf("Invalid filter selected, please select a supported AWS service")
			return
		}

		col, err := collector.NewAWSCollector()
		if err != nil {
			fmt.Errorf("Failed to create AWS collector: %v", err)
			return
		}

		// Create a map per service
		result := make(map[string]interface{})

		switch filter {
		case "ec2":
			err = collectEC2(col, result)
			if err != nil {
				return
			}
		case "rds":
			fmt.Errorf("Not implemented yet")
			return
		default:
			err := collectEC2(col, result)
			if err != nil {
				return
			}
		}
		fmt.Printf("Dumping to %s", path)
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			fmt.Errorf("Error Marshalling JSON: %v", err)
		}
		err = ioutil.WriteFile(path, jsonBytes, 0644)
		if err != nil {
			fmt.Errorf("Error writing file: %v", err)
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
		fmt.Errorf("Failed to gather EC2 Data: %v", err)
		return err
	}
	fmt.Println("Gathered %d EC2 Instances", len(instances))
	result["ec2"] = instances
	return nil
}

func init() {
	dumpCmd.AddCommand(awsCmd)
}
