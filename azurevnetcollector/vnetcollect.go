package azurevnetcollector

import (
        "context"
        "fmt"
        "github.com/adobe/cloudinventory/azurenetwork"
        "github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/network/mgmt/network"
        "sync"
        "time"
)

// AzureCollector is a struct that contains a map of subscription name and its subscriptionID
type AzureCollector struct {
        SubscriptionMap map[string]string
}

// NewAzureCollector returns an AzureCollector with subscription info set in subscriptionMap.
func NewAzureCollector() (AzureCollector, error) {
        var col AzureCollector
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
        err := col.GetSubscription(ctx)
        if err != nil {
                return col, err
        }
        return col, nil
}

// NewAzureCollectorUserDefined returns an AzureCollector with subscription info given by user in subscriptionMap.
func NewAzureCollectorUserDefined(subscriptionID []string) (AzureCollector, error) {
        var col AzureCollector
        subID := make(map[string]string)
        for i := 0; i < len(subscriptionID); i++ {
                subID["subscription_id :"+subscriptionID[i]] = subscriptionID[i]
        }
        col.SubscriptionMap = subID
        return col, nil
}

// GetSubscription adds map of subscription name with subscriptionID to the AzureCollector
func (col *AzureCollector) GetSubscription(ctx context.Context) error {
        subscription := make(map[string]string)
        var err error
        subscription, err = azurenetwork.GetAllSubscriptionIDs(ctx)
        if err != nil {
                return err
        }
        col.SubscriptionMap = subscription
        return nil

}

// CollectVirtualNetworks gathers Virtual Networks stats and data for each subscriptionID in an account level
// Function takes no.of goroutines to be created from user as input
func (col *AzureCollector) CollectVirtualNetworks(maxGoRoutines int) (map[string][]*network.VirtualNetwork, map[string]int, error) {
        VNet := make(map[string][]*network.VirtualNetwork)
        VNetCount := make(map[string]int)
        type VirtualNetworksPerSubscriptionID struct {
                SubscriptionName string
                VnetList         []*network.VirtualNetwork
        }
        var chanCapacity int

        if maxGoRoutines >= len(col.SubscriptionMap) || maxGoRoutines < 0 {
                chanCapacity = len(col.SubscriptionMap)
        } else {
                chanCapacity = maxGoRoutines
        }
        subscriptionCount := 0
        vnetsChan := make(chan VirtualNetworksPerSubscriptionID, chanCapacity)
        errChan := make(chan error, chanCapacity)
        var wg sync.WaitGroup
        for subscriptionName, subID := range col.SubscriptionMap {
                if subscriptionCount < chanCapacity {
                        wg.Add(1)
                        go func(subID string, subscriptionName string, vnetsChan chan VirtualNetworksPerSubscriptionID, errChan chan error) {
                                defer wg.Done()
                                vnets, err := CollectVirtualNetworksPerSubscriptionID(subID)
                                if err != nil {
                                        errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", subscriptionName, err))
                                        return
                                }
                                if vnets == nil {
                                        return
                                }
                                vnetsChan <- VirtualNetworksPerSubscriptionID{subscriptionName, vnets}
                        }(subID, subscriptionName, vnetsChan, errChan)
                        if subscriptionCount == chanCapacity-1 {
                                wg.Wait()
                                close(vnetsChan)
                                close(errChan)
                                if len(errChan) > 0 {
                                        return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather virtual network data: %v", <-errChan))
                                }
                                for subscriptionVnets := range vnetsChan {
                                        VNet[subscriptionVnets.SubscriptionName] = subscriptionVnets.VnetList
                                        VNetCount[subscriptionVnets.SubscriptionName] = len(subscriptionVnets.VnetList)
                                }
                        }
                } else {
                        vnets, err := CollectVirtualNetworksPerSubscriptionID(subID)
                        if err != nil {
                                return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather virtual network data: %v", err))
                        }
                        if vnets == nil {
                                subscriptionCount++
                                continue
                        }
                        VNet[subscriptionName] = vnets
                        VNetCount[subscriptionName] = len(vnets)
                }
                subscriptionCount++
        }

        return VNet, VNetCount, nil
}

// CollectVirtualNetworksPerSubscriptionID returns a slice of virtual networks for a given subscriptionID
func CollectVirtualNetworksPerSubscriptionID(subscriptionID string) ([]*network.VirtualNetwork, error) {
        vnetlist, err := azurenetwork.GetAllVNet(subscriptionID)
        return vnetlist, err
}
