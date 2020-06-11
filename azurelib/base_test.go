package azurelib

import (
	"context"
	"testing"
	"time"
)


//TestGetallSubscriptionIDs tests function GetallSubscriptionIDs
func TestGetallSubscriptionIDs(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	subscriptionmap,err := GetallSubscriptionIDs(ctx)
	if err != nil {
			t.Errorf("Failed to  Get all SubscriptionIDs: %v", err)
	} else {
			t.Logf("GetallSubscriptionIDs successful.Found %d subscription IDs",len(subscriptionmap))
			
	}
}
