package azurelib

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-05-01/resources"
)


// Tests if the resource describer is working
func TestDescribeResource(t *testing.T) {
	subscriptionID := "282160c0-3c83-43f1-bff1-9356b1678ffb"

	sess, _ := newSession()

	grClient := resources.NewGroupsClient(subscriptionID)
	grClient.Authorizer = sess.Authorizer

	for list, err := grClient.ListComplete(context.Background(), "", nil); list.NotDone(); err = list.Next() {
		if err != nil {
			t.Errorf("resources cannot be listed. Error: %s \n", err)
		}

		rgID := *list.Value().ID
		_, rgErr := describeResource(sess, subscriptionID, rgID)

		if rgErr != nil {
			t.Errorf("cannot describe the resource %s. Error - %s \n", rgID, rgErr)
			break
		}
	}

}
