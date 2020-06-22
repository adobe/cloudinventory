package azurelib

import (
        "context"
        "github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/network/mgmt/network"
        "github.com/Azure/go-autorest/autorest/azure/auth"
        "time"
)

// GetAllLdb function returns a list of load balancers for a given subscriptionID
func GetAllLdb(subscriptionID string) ([]*network.LoadBalancer, error) {
        var ldbList []*network.LoadBalancer
        authorizer, err := auth.NewAuthorizerFromEnvironment()
        if err != nil {
                return ldbList, err
        }
        ldbClient := network.NewLoadBalancersClient(subscriptionID)
        ldbClient.Authorizer = authorizer
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
        result, err := ldbClient.ListAllComplete(ctx)
        if err != nil {
                return ldbList, err
        }
        for result.NotDone() {
                ldb := result.Value()
                ldbList = append(ldbList, &ldb)
                if err = result.Next(); err != nil {
                        return ldbList, err
                }
        }

        return ldbList, nil
}