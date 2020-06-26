package azurenetwork

import (
        "context"
        "testing"
        "time"
)

// TestGetAllVNet tests the function GetAllVNet
func TestGetAllVNet(t *testing.T) {
        if testing.Short() {
                t.Skip("Skipping test in short mode")
        }
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
        subscriptions, err := GetAllSubscriptionIDs(ctx)
        if err != nil {
                t.Errorf("Unable to get subscriptionIDs: %v", err)
        }
        for key, subsID := range subscriptions {
                Dbs, err := GetAllVNet(subsID)
                if err != nil {
                        t.Errorf("Failed to get virtual networks for subscription: %s because %v", key, err)
                }
                t.Logf("Found %d virtual networks in %s", len(Dbs), key)
        }
}
