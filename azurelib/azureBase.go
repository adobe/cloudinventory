package azurelib

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/preview/preview/subscription/mgmt/subscription"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

//AzureSession - struct that holds the Azure Session. Goal is to pass it to multiple methods that require authorization
type AzureSession struct {
	//SubscriptionID string
	Authorizer autorest.Authorizer
}

// makes use of env vars and creates a new session
func newSession() (*AzureSession, error) {

	envVars := []string{"AZURE_TENANT_ID", "AZURE_CLIENT_ID", "AZURE_CLIENT_SECRET"}

	for _, eachvar := range envVars {
		if len(os.Getenv(eachvar)) == 0 {
			return nil, fmt.Errorf("%s is empty", eachvar)
		}
	}
	authorizer, err := auth.NewAuthorizerFromEnvironment()

	if err != nil {
		fmt.Println("error from session init", err)
		return nil, err
	}

	sess := AzureSession{
		Authorizer: authorizer,
	}
	return &sess, nil
}

// returns a list of subscriptions that the client_id has access to
func ListSubscriptions(sess *AzureSession) (listOfSubscriptions []string, err error) {

	//var listOfSubscriptions []string

	subClient := subscription.NewSubscriptionsClient()
	subClient.Authorizer = sess.Authorizer

	for subList, subErr := subClient.List(context.Background()); subList.NotDone(); subErr = subList.Next() {
		if subErr != nil {
			fmt.Println("subscription list cannot be obtained")
			return nil, subErr
		}
		for _, eachSub := range subList.Values() {
			listOfSubscriptions = append(listOfSubscriptions, strings.Split(string(*eachSub.ID), "/")[2])
		}

	}
	return listOfSubscriptions, nil
}
