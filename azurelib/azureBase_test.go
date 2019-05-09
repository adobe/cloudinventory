package azurelib

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-05-01/resources"
)

// Tests if an authorizer could be obtained from env props and if a session gets created
func TestNewSession(t *testing.T) {

	sess, err := newSession()
	listOfSubscriptions, _ := ListSubscriptions(sess)

	for _, eachSub := range listOfSubscriptions {
		if err != nil {
			t.Errorf("authorizer could not be created. Error: %s", err)
		}
		resourcesClient := resources.NewGroupsClient(eachSub)
		resourcesClient.Authorizer = sess.Authorizer

		_, resErr := resourcesClient.ListComplete(context.Background(), "", nil)

		if resErr != nil {
			t.Errorf("session cannot be obtained. Here is the cliendID: %s \n", os.Getenv("AZURE_CLIENT_ID"))
		}
	}
}

//Tests if list of subscriptions could be obtained
func TestListSubscriptions(t *testing.T) {

	envVars := []string{"AZURE_TENANT_ID", "AZURE_CLIENT_ID", "AZURE_CLIENT_SECRET"}

	for _, eachvar := range envVars {
		if len(os.Getenv(eachvar)) == 0 {
			t.Errorf("%s is empty", eachvar)
		}
	}

	sess, _ := newSession()

	listOfSubscripions, subErr := ListSubscriptions(sess)

	if subErr != nil {
		t.Errorf("Cannot list the subscriptions. Here is the error: %s", subErr)
	}

	if len(listOfSubscripions) == 0 {
		t.Errorf("Cannot list the subscriptions. Here is the cliendID: %s \n", os.Getenv("AZURE_CLIENT_ID"))
	}
}
