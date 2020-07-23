package azurelib

import (
        "context"
        "github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/resources/mgmt/subscriptions"
        "github.com/Azure/go-autorest/autorest/azure/auth"
)

// GetAllSubscriptionIDs function returns a map of subscription name and subscription ID at account level
func GetAllSubscriptionIDs(ctx context.Context) (map[string]string, error) {
        subscriptionMap := make(map[string]string)
        authorizer, err := auth.NewAuthorizerFromEnvironment()
        if err != nil {
                return subscriptionMap, err
        }
        newClient := subscriptions.NewClient()
        newClient.Authorizer = authorizer
        result, err := newClient.ListComplete(ctx)
        if err != nil {
                return subscriptionMap, err
        }
        for result.NotDone() {
                subscription := result.Value()
                subscriptionMap[*subscription.DisplayName] = *subscription.SubscriptionID
                if err := result.Next(); err != nil {
                        return subscriptionMap, err
                }
        }
        return subscriptionMap, nil

}
