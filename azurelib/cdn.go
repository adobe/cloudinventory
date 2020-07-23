package azurelib

import (
        "context"
	"github.com/Azure/azure-sdk-for-go/services/cdn/mgmt/2019-04-15/cdn"
        "github.com/Azure/go-autorest/autorest/azure/auth"
        "strings"
        "time"
)

// GetAllCDN function returns a list of CDN instances for a given subscriptionID
func GetAllCDN(subscriptionID string) ([]*cdn.Endpoint, error) {
        var cdnList []*cdn.Endpoint  
        authorizer, err := auth.NewAuthorizerFromEnvironment()
        if err != nil {
                return cdnList, err
        }
        cdnClient := cdn.NewProfilesClient(subscriptionID)
        endPointClient:=cdn.NewEndpointsClient(subscriptionID)

        cdnClient.Authorizer = authorizer
        endPointClient.Authorizer = authorizer
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
        result, err := cdnClient.ListComplete(ctx)
        if err != nil {
                return cdnList, err
        }
        for result.NotDone() {
                cdn := result.Value()
                profile := *cdn.Name
                cdnID := *cdn.ID
                ID := strings.Split(cdnID,"/")
                resourceGroup := ID[4]
                endpointResult,err := endPointClient.ListByProfileComplete(ctx, resourceGroup, profile)
                if err != nil {
                        return cdnList, err
                }
                for endpointResult.NotDone(){
                        endpoint := endpointResult.Value()
                        cdnList = append(cdnList, &endpoint)
                        if err = endpointResult.Next(); err != nil {
                                return cdnList, err
                        }
                }
                if err = result.Next(); err != nil {
                        return cdnList, err
                }
                
        }

        return cdnList, nil
}
