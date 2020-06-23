package cmd

import (
        "encoding/json"
        "fmt"
        "github.com/adobe/cloudinventory/azurecollector"
        "github.com/spf13/cobra"
        "io/ioutil"
)

// azureCmd represents the azure command
var azureCmd = &cobra.Command{
        Use:   "azure",
        Short: "Dump Azure inventory. Currently supports Virtual Machines/SQL databases",
        Run: func(cmd *cobra.Command, args []string) {
                path := cmd.Flag("path").Value.String()
                filter := cmd.Flag("filter").Value.String()
                if !validateAzureFilter(filter) {
                        fmt.Printf("Invalid filter selected, please select a supported Azure service")
                        return
                }

                col, err := azurecollector.NewAzureCollector()
                if err != nil {
                        fmt.Printf("Failed to create Azure collector: %v\n", err)
                        return
                }

                // Create a map per service
                result := make(map[string]interface{})

                switch filter {
                case "vm":
                        err := collectVMS(col, result)
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
                        err := collectVMS(col, result)
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

func collectVMS(col azurecollector.AzureCollector, result map[string]interface{}) error {
        instances, err := col.CollectVMS()
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
        result["ldb"] = instances
        return nil
}

func init() {
        dumpCmd.AddCommand(azureCmd)
}
