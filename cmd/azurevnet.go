package cmd

import (
        "encoding/json"
        "fmt"
        "github.com/adobe/cloudinventory/azurevnetcollector"
        "github.com/spf13/cobra"
        "io/ioutil"
        "strings"
)

// azurevnetCmd represents the azure command
var azurevnetCmd = &cobra.Command{
        Use:   "azurevnet",
        Short: "Dump Azure inventory. Currently supports Virtual networks",
        Run: func(cmd *cobra.Command, args []string) {
                path := cmd.Flag("path").Value.String()
                input_Path := cmd.Flag("input_Path").Value.String()
                statistics, _ := cmd.Flags().GetBool("stats")
                GoRoutines, _ := cmd.Flags().GetInt("maxGoRoutines")
                var col azurevnetcollector.AzureCollector
                var err error
                if input_Path != "" {
                        data, err := ioutil.ReadFile(input_Path)
                        if err != nil {
                                fmt.Println("File reading error", err)
                                return
                        }
                        s := string(data)
                        subID := strings.Split(s, " ")
                        col, err = azurevnetcollector.NewAzureCollectorUserDefined(subID)
                        if err != nil {
                                fmt.Printf("Failed to create Azure vnet collector: %v\n", err)
                                return
                        }

                } else {
                        col, err = azurevnetcollector.NewAzureCollector()
                        if err != nil {
                                fmt.Printf("Failed to create Azure vnet collector: %v\n", err)
                                return
                        }
                }
                
                // Create a map per service
                result := make(map[string]interface{})
                if statistics {
                        err = collectVNetStats(col, GoRoutines, result)
		        if err != nil {
                                return
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
		        err = collectVNets(col, GoRoutines, result)
		        if err != nil {
                                return
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

func collectVNets(col azurevnetcollector.AzureCollector, maxGoRoutines int, result map[string]interface{}) error {
        instances, _, err := col.CollectVirtualNetworks(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather virtual networks Data: %v\n", err)
                return err
        }
        fmt.Printf("Gathered virtual networks across %d subscriptions\n", len(instances))
        result["vnet"] = instances
        return nil
}

func collectVNetStats(col azurevnetcollector.AzureCollector, maxGoRoutines int, resultStats map[string]interface{}) error {
        _, stats, err := col.CollectVirtualNetworks(maxGoRoutines)
        if err != nil {
                fmt.Printf("Failed to gather virtual networks stats: %v\n", err)
                return err
        }
        fmt.Printf("Gathered virtual networks stats across %d subscriptions\n", len(stats))
        resultStats["vnet"] = stats
        return nil
}

func init() {
        dumpCmd.AddCommand(azurevnetCmd)
        azurevnetCmd.PersistentFlags().StringP("input_Path", "i", "", "file path to take subscriptionIDs as input")
        azurevnetCmd.PersistentFlags().BoolP("stats", "s", false, "dumps the stats of different resources for subscriptions")
        azurevnetCmd.PersistentFlags().IntP( "maxGoRoutines","m", -1, "customize maximum no.of Goroutines ")
}
