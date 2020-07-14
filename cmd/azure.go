package cmd

import (
        "encoding/json"
        "fmt"
        "github.com/adobe/cloudinventory/azurecollector"
        "github.com/spf13/cobra"
        "io/ioutil"
        "strings"
)

var maxGoRoutines int

// azureCmd represents the azure command
var azureCmd = &cobra.Command{
        Use:   "azure",
        Short: "Dump Azure inventory. Currently supports Virtual Machines/SQL databases/Load balancers",
        Run: func(cmd *cobra.Command, args []string) {
                path := cmd.Flag("path").Value.String()
                filter := cmd.Flag("filter").Value.String()
                inputPath := cmd.Flag("inputPath").Value.String()
                if !validateAzureFilter(filter) {
                        fmt.Printf("Invalid filter selected, please select a supported Azure service")
                        return
                }
                var col azurecollector.AzureCollector
                var err error
                if inputPath != "" {
                        data, err := ioutil.ReadFile(inputPath)
                        if err != nil {
                                fmt.Println("File reading error", err)
                                return
                        }
                        s := string(data)
                        subID := strings.Split(s, " ")
                        col, err = azurecollector.NewAzureCollectorUserDefined(subID)
                        if err != nil {
                                fmt.Printf("Failed to create Azure collector: %v\n", err)
                                return
                        }

                } else {
                        col, err = azurecollector.NewAzureCollector()
                        if err != nil {
                                fmt.Printf("Failed to create Azure collector: %v\n", err)
                                return
                        }
                }

                if maxGoRoutines <= -1 {
                        maxGoRoutines = len(col.SubscriptionMap)
                }
                // Create a map per service
                result := make(map[string]interface{})

                switch filter {
                case "vm":
                        err := collectVMS(col, maxGoRoutines, result)
                        if err != nil {
                                return
                        }
                case "sqldb":
                        err := collectSQLDB(col, result)
                        if err != nil {
                                return
                        }
                case "loadbalancer":
                        err := collectLDB(col, result)
                        if err != nil {
                                return
                        }
                default:
                        err := collectVMS(col, maxGoRoutines, result)
                        if err != nil {
                                return
                        }
                        err = collectSQLDB(col, result)
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

        },
}

func validateAzureFilter(filter string) bool {
        validSlice := []string{
                "vm",
                "sqldb",
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

func collectVMS(col azurecollector.AzureCollector, maxGoRoutines int, result map[string]interface{}) error {
        instances, err := col.CollectVMS(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather VM Data: %v\n", err)
                return err
        }
        fmt.Printf("Gathered VM Instances across %d subscriptions\n", len(instances))
        result["vm"] = instances
        return nil
}

func collectSQLDB(col azurecollector.AzureCollector, result map[string]interface{}) error {
        instances, err := col.CollectSQLDBs()
        if err != nil {
                fmt.Printf("Failed to gather SQL database Data: %v\n", err)
                return err
        }
        fmt.Printf("Gathered SQL databases across %d subscriptions\n", len(instances))
        result["sqldb"] = instances
        return nil
}

func collectLDB(col azurecollector.AzureCollector, result map[string]interface{}) error {
        instances, err := col.CollectLoadBalancers()
        if err != nil {
                fmt.Printf("Failed to gather load balancers Data: %v\n", err)
                return err
        }
        fmt.Printf("Gathered load balancers across %d subscriptions\n", len(instances))
        result["loadbalancer"] = instances
        return nil
}

func init() {
        dumpCmd.AddCommand(azureCmd)
        azureCmd.PersistentFlags().StringP("inputPath", "i", "", "file path to take subscriptionIDs as input")
        azureCmd.PersistentFlags().IntVarP(&maxGoRoutines, "maxGoRoutines","m", -1, "customize maximum no.of Goroutines ")
}
