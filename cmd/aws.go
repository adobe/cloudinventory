package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"github.com/tchaudhry91/cloudinventory/ansible"
	"github.com/tchaudhry91/cloudinventory/collector"
)

var partition string
var ansibleinv string
var ansibleEnable bool
var ansiblePriv bool

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

		col, err := collector.NewAWSCollector(partition, nil)
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
		fmt.Printf("Dumping to %s\n", path)
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			fmt.Printf("Error Marshalling JSON: %v\n", err)
		}
		err = ioutil.WriteFile(path, jsonBytes, 0644)
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
		}

		if ansibleEnable {
			fmt.Printf("Building Inventory for Ansible at: %s", ansibleinv)
			ansinv, err := ansible.BuildEC2Inventory(result["ec2"].(map[string][]*ec2.Instance), ansiblePriv)
			if err != nil {
				fmt.Printf("Error while building Ansible Inventory: %v\n", err)
			}
			err = ioutil.WriteFile(ansibleinv, []byte(ansinv), 0644)
			if err != nil {
				fmt.Printf("Error writing to Ansible Inventory file: %v\n", err)
			}
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
	awsCmd.PersistentFlags().StringVarP(&partition, "partition", "", "default", "Which partition of AWS to run for default/china")
	awsCmd.PersistentFlags().BoolVarP(&ansibleEnable, "ansible", "a", false, "Create a an ansible inventory as well (only for EC2)")
	awsCmd.PersistentFlags().StringVarP(&ansibleinv, "ansible_inv", "", "ansible.inv", "File to create the EC2 ansible inventory in")
	awsCmd.PersistentFlags().BoolVarP(&ansiblePriv, "ansible_private", "", false, "Create Ansible Inventory with private DNS instead of public")
	dumpCmd.AddCommand(awsCmd)
}
