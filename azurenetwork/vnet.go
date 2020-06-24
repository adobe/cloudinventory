package azurenetwork

import (
        "context"
        "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-10-01/network"
        "github.com/Azure/go-autorest/autorest/azure/auth"
        "time"
)

// GetAllVNet function gathers all virtual networks for a given subscriptionID
func GetAllVNet(subscriptionID string) ([]*network.VirtualNetwork, error) {
        var vnetList []*network.VirtualNetwork
        authorizer, err := auth.NewAuthorizerFromEnvironment()
        if err != nil {
                return vnetList, err
        }
        vnetClient := network.NewVirtualNetworksClient(subscriptionID)
        vnetClient.Authorizer = authorizer
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
        results, err := vnetClient.ListAllComplete(ctx)
        if err != nil {
                return vnetList, err
        }
        for results.NotDone() {
                vnet := results.Value()
                vnetList = append(vnetList, &vnet)
                if err := results.Next(); err != nil {
                        return vnetList, err
                }

        }

        return vnetList, nil

}
