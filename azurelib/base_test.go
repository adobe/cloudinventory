package azurelib

import (
        "context"
        "testing"
        "time"
)

// TestGetAllSubscriptionIDs tests function GetAllSubscriptionIDs
func TestGetAllSubscriptionIDs(t *testing.T) {

        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
        subscriptionMap, err := GetAllSubscriptionIDs(ctx)
        if err != nil {
                t.Errorf("Failed to  Get all SubscriptionIDs: %v", err)
        } else {
                t.Logf("GetallSubscriptionIDs successful.Found %d subscription IDs", len(subscriptionMap))

        }
}
