package azurelib

import (
        "context"
        "testing"
        "time"
)

// AuthorizeClients function creates clients and authorizes all the clients
func GetAuthorizedClients(subscriptionID string) (client Clients, err error) {
        client = GetNewClients(subscriptionID)
        err = client.AuthorizeClients()
        return
}

// TestGetAllVMS tests function GetallVMS
func TestGetAllVMS(t *testing.T) {
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
                client, err := GetAuthorizedClients(subsID)
                if err != nil {
                        t.Errorf("Failed to get virtual machines for subscription: %s because %v", key, err)
                }
                Vmlist, err := GetAllVMS(client)
                if err != nil {
                        t.Errorf("Failed to get virtual machines for subscription: %s because %v", key, err)
                }
                t.Logf("Found %d virtual machines in %s", len(Vmlist), key)

        }
}

// TestGetAllSQLDBs tests the function GetallSQLDBs
func TestGetAllSQLDBs(t *testing.T) {
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
                Dbs, err := GetAllSQLDBs(subsID)
                if err != nil {
                        t.Errorf("Failed to get databases for subscription: %s because %v", key, err)
                }
                t.Logf("Found %d databases in %s", len(Dbs), key)
        }
}
