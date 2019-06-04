package azurelib

import (
	"context"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/profiles/preview/preview/subscription/mgmt/subscription"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

func init() {

	var logFile string = "azureInventory.log"
	//create a log file if it does not exist. Otherwise, append to the existing file
	fPtr, _ := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

	// log JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	log.SetOutput(fPtr)
	log.SetLevel(log.InfoLevel)

}

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
			log.Warning("Azure credentials cannot be obtained from the environment")
			return nil, fmt.Errorf("%s is empty", eachvar)
		}
	}
	authorizer, err := auth.NewAuthorizerFromEnvironment()

	if err != nil {
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
			log.Warning("subscription list cannot be obtained")
			return nil, subErr
		}
		for _, eachSub := range subList.Values() {
			listOfSubscriptions = append(listOfSubscriptions, strings.Split(string(*eachSub.ID), "/")[2])
		}

	}
	return listOfSubscriptions, nil
}
