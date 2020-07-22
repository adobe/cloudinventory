package cmd

import (
        "encoding/json"
        "fmt"
        "github.com/adobe/cloudinventory/azurecollector"
        "github.com/spf13/cobra"
        "io/ioutil"
        "strings"
)

// azureCmd represents the azure command
var azureCmd = &cobra.Command{
        Use:   "azure",
        Short: "Dump Azure inventory. Currently supports Virtual Machines/SQL databases/Load balancers/CDN",
        Run: func(cmd *cobra.Command, args []string) {
                path := cmd.Flag("path").Value.String()
                filter := cmd.Flag("filter").Value.String()
                inputPath := cmd.Flag("inputPath").Value.String()
                statistics, _ := cmd.Flags().GetBool("stats")
                maxGoRoutines, _ := cmd.Flags().GetInt("maxGoRoutines")
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

                // Create a map per service
                result := make(map[string]interface{})
                
                if statistics {
                        switch filter {
                        case "vm":
                                err := collectVMSStats(col, maxGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "sqldb":
                                err := collectSQLDBStats(col, maxGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "loadbalancer":
                                err := collectLDBStats(col, maxGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "cdn":
                                err := collectCDNStats(col, maxGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        default:
                                err := collectVMSStats(col, maxGoRoutines, result)
                                if err != nil {
                                        return
                                }
                                err = collectSQLDBStats(col, maxGoRoutines, result)
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
                        case "vm":
                                err := collectVMS(col, maxGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "sqldb":
                                err := collectSQLDB(col, maxGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "loadbalancer":
                                err := collectLDB(col, maxGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        case "cdn":
                                err := collectCDN(col, maxGoRoutines, result)
                                if err != nil {
                                        return
                                }
                        default:
                                err := collectVMS(col, maxGoRoutines, result)
                                if err != nil {
                                        return
                                }
                                err = collectSQLDB(col, maxGoRoutines, result)
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
               
        },
}

func validateAzureFilter(filter string) bool {
        validSlice := []string{
                "vm",
                "sqldb",
                "loadbalancer",
                "cdn",
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

func collectSQLDB(col azurecollector.AzureCollector, maxGoRoutines int, result map[string]interface{}) error {
        instances, err := col.CollectSQLDBs(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather SQL database Data: %v\n", err)
                return err
        }
        fmt.Printf("Gathered SQL databases across %d subscriptions\n", len(instances))
        result["sqldb"] = instances
        return nil
}

func collectLDB(col azurecollector.AzureCollector, maxGoRoutines int, result map[string]interface{}) error {
        instances, _, err := col.CollectLoadBalancers(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather load balancers Data: %v\n", err)
                return err
        }
        fmt.Printf("Gathered load balancers across %d subscriptions\n", len(instances))
        result["loadbalancer"] = instances
        return nil
}

func collectCDN(col azurecollector.AzureCollector, maxGoRoutines int, result map[string]interface{}) error {
        instances, _, err := col.CollectCDN(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather CDN Data: %v\n", err)
                return err
        }
        fmt.Printf("Gathered CDN across %d subscriptions\n", len(instances))
        result["cdn"] = instances
        return nil
}

func collectVMSStats(col azurecollector.AzureCollector, maxGoRoutines int, resultStats map[string]interface{}) error {
        
        stats, err := col.CollectVMSCount(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather VM Stats: %v\n", err)
                return err
        }
        fmt.Printf("Gathered VM Stats across %d subscriptions\n", len(stats))
        resultStats["vm"] = stats 
        return nil
}

func collectSQLDBStats(col azurecollector.AzureCollector, maxGoRoutines int, resultStats map[string]interface{}) error {
        stats, err := col.CollectSQLDBCount(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather SQL database Stats: %v\n", err)
                return err
        }
        fmt.Printf("Gathered SQL databases Stats across %d subscriptions\n", len(stats))
        resultStats["sqldb"] = stats
        return nil
}

func collectLDBStats(col azurecollector.AzureCollector, maxGoRoutines int, resultStats map[string]interface{}) error {
       _, stats, err := col.CollectLoadBalancers(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather load balancers Stats: %v\n", err)
                return err
        }
        fmt.Printf("Gathered load balancers Stats across %d subscriptions\n", len(stats))
        resultStats["loadbalancer"] = stats
        return nil
}

func collectCDNStats(col azurecollector.AzureCollector, maxGoRoutines int, resultStats map[string]interface{}) error {
        _, stats, err := col.CollectCDN(maxGoRoutines)
         if err != nil {
                 fmt.Printf("Failed to gather CDN Stats: %v\n", err)
                 return err
         }
         fmt.Printf("Gathered CDN Stats across %d subscriptions\n", len(stats))
         resultStats["cdn"] = stats
         return nil
 }

func init() {
        dumpCmd.AddCommand(azureCmd)
        azureCmd.PersistentFlags().StringP("inputPath", "i", "", "file path to take subscriptionIDs as input")
        azureCmd.PersistentFlags().BoolP("stats", "s", false, "dumps the stats of different resources for subscriptions")
        azureCmd.PersistentFlags().IntP("maxGoRoutines","m", -1, "customize maximum no.of Goroutines ")
}
