/*
Copyright 2019 Adobe. All rights reserved.
This file is licensed to you under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License. You may obtain a copy
of the License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
OF ANY KIND, either express or implied. See the License for the specific language
governing permissions and limitations under the License.
*/

package cmd

import (
        "encoding/json"
        "fmt"
        "io/ioutil"

        "github.com/adobe/cloudinventory/ansible"
        "github.com/adobe/cloudinventory/collector"
        "github.com/aws/aws-sdk-go/service/ec2"
        "github.com/spf13/cobra"
        "strings"
)

var partition string
var ansibleinv string
var ansibleEnable bool
var ansiblePriv bool
var maximumGoRoutines int

// awsCmd represents the aws command
var awsCmd = &cobra.Command{
        Use:   "aws",
        Short: "Dump AWS inventory. Currently supports EC2/RDS/Route53/LoadBalancers/CloudFront/Vpc/Subnets",
        Run: func(cmd *cobra.Command, args []string) {
                path := cmd.Flag("path").Value.String()
                filter := cmd.Flag("filter").Value.String()
                inputPath := cmd.Flag("inputPath").Value.String()
                statistics, _ := cmd.Flags().GetBool("stats")
                maximumGoRoutines, _ := cmd.Flags().GetInt("maxGoRoutines")
                if !validateAWSFilter(filter) {
                        fmt.Printf("Invalid filter selected, please select a supported AWS service")
                        return
                }
                var col collector.AWSCollector
                var err error
                if inputPath != "" {
                        data, err := ioutil.ReadFile(inputPath)
                        if err != nil {
                                fmt.Println("File reading error", err)
                                return
                        }
                        s := string(data)
                        regions := strings.Split(s, " ")
                        col, err = collector.NewAWSCollectorUserDefined(regions, nil)
                        if err != nil {
                                fmt.Printf("Failed to create AWS collector: %v\n", err)
                                return
                        }

                } else {
                        col, err = collector.NewAWSCollector(partition, nil)
                        if err != nil {
                                fmt.Printf("Failed to create AWS collector: %v\n", err)
                                return
                        }
                }

                // Create a map per service
                result := make(map[string]interface{})
                
                if statistics {
                        switch filter {
                        case "ec2":
                                err := collectEC2Stats(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "rds":
                                err := collectRDSStats(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "hostedzone":
                                err := collectHostedZoneStats(col, result)
                                if err != nil {
                                        return
                                }
                        case "loadbalancer":
                                err := collectLoadBalancerStats(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "cloudfront":
                                err := collectCloudFrontStats(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "vpc":
                                err := collectVpcStats(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "subnet":
                                err := collectSubnetStats(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }                
                        default:
                                err := collectEC2Stats(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                                err = collectRDSStats(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        }
                        fmt.Printf("Dumping stats to %s\n", path)
                        jsonBytes, err := json.MarshalIndent(result, "", "    ")
                        if err != nil {
                                fmt.Printf("Error Marshalling JSON: %v\n", err)
                        }
                        err = ioutil.WriteFile(path, jsonBytes, 0644)
                        if err != nil {
                                fmt.Printf("Error writing file: %v\n", err)
                        }
                } else {
                        switch filter {
                        case "ec2":
                                err := collectEC2(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "rds":
                                err := collectRDS(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "hostedzone":
                                err := collectHostedZone(col, result)
                                if err != nil {
                                        return
                                }
                        case "loadbalancer":
                                err := collectLoadBalancers(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "cloudfront":
                                err := collectCloudFront(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "vpc":
                                err := collectVpc(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "subnet":
                                err := collectSubnets(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }                
                        default:
                                err := collectEC2(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                                err = collectRDS(col, maximumGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        }
                        fmt.Printf("Dumping to %s\n", path)
                        jsonBytes, err := json.MarshalIndent(result, "", "    ")
                        if err != nil {
                                fmt.Printf("Error Marshalling JSON: %v\n", err)
                        }
                        err = ioutil.WriteFile(path, jsonBytes, 0644)
                        if err != nil {
                                fmt.Printf("Error writing file: %v\n", err)
                        }
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
                "cloudfront",
                "vpc",
                "subnet",
                "rds",
                "hostedzone",
                "loadbalancer",
                "",
        }
        for _, service := range validSlice {
                if filter == service {
                        return true
                }
        }
        return false
}

func collectEC2(col collector.AWSCollector, maxGoRoutines int, result map[string]interface{}) error {
        instances, _, err := col.CollectEC2(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather EC2 Data: %v\n", err)
                return err
        }
        fmt.Printf("Gathered EC2 Instances across %d regions\n", len(instances))
        result["ec2"] = instances
        return nil
}

func collectHostedZone(col collector.AWSCollector, result map[string]interface{}) error {
        instances, _, err := col.CollectZones()
        if err != nil {
                fmt.Printf("Failed to gather HostedZones Data: %v\n", err)
                return err
        }
        fmt.Printf("Gathered HostedZone data across all regions\n")
        result["hostedzones"] = instances
        return nil
}

func collectVpc(col collector.AWSCollector, maxGoRoutines int, result map[string]interface{}) error {
        instances, _, err := col.CollectVPC(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather Vpc Data: %v\n", err)
                return err
        }
        fmt.Printf("Gathered Vpc Instances across %d regions\n", len(instances))
        result["vpc"] = instances
        return nil
}

func collectSubnets(col collector.AWSCollector, maxGoRoutines int, result map[string]interface{}) error {
        instances, _, err := col.CollectSubnets(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather subnets Data: %v\n", err)
                return err
        }
        fmt.Printf("Gathered Subnet Instances across %d regions\n", len(instances))
        result["subnet"] = instances
        return nil
}

func collectCloudFront(col collector.AWSCollector, maxGoRoutines int, result map[string]interface{}) error {
        instances, _, err := col.CollectCloudFront(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather cloudfront Data: %v\n", err)
                return err
        }
        fmt.Printf("Gathered cloudfront Instances across %d regions\n", len(instances))
        result["cdn"] = instances
        return nil
}

func collectRDS(col collector.AWSCollector, maxGoRoutines int, result map[string]interface{}) error {
        instances, _, err := col.CollectRDS(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather RDS Data: %v\n", err)
                return err
        }
        fmt.Printf("Gathered RDS Instances across %d regions\n", len(instances))
        result["rds"] = instances
        return nil
}

func collectLoadBalancers(col collector.AWSCollector, maxGoRoutines int, result map[string]interface{}) error {

        var allLbs []interface{}
        clbs, _, err := col.CollectClassicLoadBalancers(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather classic load balancers: %v\n", err)
                return err
        }
        allLbs = append(allLbs, clbs)
        fmt.Printf("Gathered Classic Load Balancers across %d regions\n", len(clbs))

        anlbs, _, err := col.CollectApplicationAndNetworkLoadBalancers(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather application and network load balancers: %v\n", err)
                return err
        }
        allLbs = append(allLbs, anlbs)
        fmt.Printf("Gathered Application and Network Load Balancers across %d regions\n", len(anlbs))

        result["loadbalancer"] = allLbs
        return nil
}


func collectEC2Stats(col collector.AWSCollector, maxGoRoutines int, result map[string]interface{}) error {
        _, stats, err := col.CollectEC2(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather EC2 stats: %v\n", err)
                return err
        }
        fmt.Printf("Gathered EC2 Instance stats across %d regions\n", len(stats))
        result["ec2"] = stats
        return nil
}

func collectHostedZoneStats(col collector.AWSCollector, result map[string]interface{}) error {
        _, stats, err := col.CollectZones()
        if err != nil {
                fmt.Printf("Failed to gather HostedZones stats: %v\n", err)
                return err
        }
        fmt.Printf("Gathered HostedZone stats across all regions\n")
        result["hostedzones"] = stats
        return nil
}

func collectVpcStats(col collector.AWSCollector, maxGoRoutines int, result map[string]interface{}) error {
        _, stats, err := col.CollectVPC(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather Vpc stats: %v\n", err)
                return err
        }
        fmt.Printf("Gathered Vpc stats across %d regions\n", len(stats))
        result["vpc"] = stats
        return nil
}

func collectSubnetStats(col collector.AWSCollector, maxGoRoutines int, result map[string]interface{}) error {
        _, stats, err := col.CollectSubnets(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather subnets stats: %v\n", err)
                return err
        }
        fmt.Printf("Gathered Subnet stats across %d regions\n", len(stats))
        result["subnet"] = stats
        return nil
}

func collectCloudFrontStats(col collector.AWSCollector, maxGoRoutines int, result map[string]interface{}) error {
        _, stats, err := col.CollectCloudFront(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather cloudfront stats: %v\n", err)
                return err
        }
        fmt.Printf("Gathered cloudfront stats across %d regions\n", len(stats))
        result["cdn"] = stats
        return nil
}

func collectRDSStats(col collector.AWSCollector, maxGoRoutines int, result map[string]interface{}) error {
        _, stats, err := col.CollectRDS(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather RDS stats: %v\n", err)
                return err
        }
        fmt.Printf("Gathered RDS Instances stats across %d regions\n", len(stats))
        result["rds"] = stats
        return nil
}

func collectLoadBalancerStats(col collector.AWSCollector, maxGoRoutines int, result map[string]interface{}) error {

        var allLbstats []interface{}
        _, clbStats, err := col.CollectClassicLoadBalancers(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather classic load balancers stats: %v\n", err)
                return err
        }
        allLbstats = append(allLbstats, clbStats)
        fmt.Printf("Gathered Classic Load Balancers stats across %d regions\n", len(clbStats))

        _, albStats, err := col.CollectApplicationAndNetworkLoadBalancers(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather application and network load balancers stats: %v\n", err)
                return err
        }
        allLbstats = append(allLbstats, albStats)
        fmt.Printf("Gathered Application and Network Load Balancers stats across %d regions\n", len(albStats))

        result["loadbalancer"] = allLbstats
        return nil
}

func init() {
        awsCmd.PersistentFlags().StringVarP(&partition, "partition", "", "default", "Which partition of AWS to run for default/china")
        awsCmd.PersistentFlags().BoolVarP(&ansibleEnable, "ansible", "a", false, "Create a an ansible inventory as well (only for EC2)")
        awsCmd.PersistentFlags().StringVarP(&ansibleinv, "ansible_inv", "", "ansible.inv", "File to create the EC2 ansible inventory in")
        awsCmd.PersistentFlags().BoolVarP(&ansiblePriv, "ansible_private", "", false, "Create Ansible Inventory with private DNS instead of public")
        awsCmd.PersistentFlags().IntP("maxGoRoutines","m", -1, "customize maximum no.of Goroutines ")
        awsCmd.PersistentFlags().BoolP("stats", "s", false, "dumps the stats of different resources for regions")
        awsCmd.PersistentFlags().StringP("inputPath", "i", "", "file path to take user input")
        dumpCmd.AddCommand(awsCmd)
}
