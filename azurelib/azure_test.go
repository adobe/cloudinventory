package azurelib

import (
        "context"
        "testing"
        "time"
)

//AuthorizeClients function creates clients and authorizes all the clients
func GetAuthorizedclients(subscriptionID string) (client Clients, err error) {
        clients := GetNewClients(subscriptionID)
        client, err = AuthorizeClients(clients)
        return
}

//TestGetallVMS tests function GetallVMS
func TestGetallVMS(t *testing.T) {
        if testing.Short() {
                t.Skip("Skipping test in short mode")
        }
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()

        subscriptions, err := GetallSubscriptionIDs(ctx)
        if err != nil {
                t.Errorf("Unable to get subscriptionIDs: %v", err)
        }

        for key, subsID := range subscriptions {
                client, err := GetAuthorizedclients(subsID)
                if err != nil {

                        t.Errorf("Failed to get virtual machines for subscription: %s because %v", key, err)
                }

                Vmlist, err := GetallVMS(client, ctx)
                if err != nil {
                        t.Errorf("Failed to get virtual machines for subscription: %s because %v", key, err)
                }
                t.Logf("Found %d virtual machines in %s", len(Vmlist), key)

        }
}

//TestGetallSQLDBs tests the function GetallSQLDBs
func TestGetallSQLDBs(t *testing.T) {
        if testing.Short() {
                t.Skip("Skipping test in short mode")
        }
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
        subscriptions, err := GetallSubscriptionIDs(ctx)
        if err != nil {
                t.Errorf("Unable to get subscriptionIDs: %v", err)
        }
        for key, subsID := range subscriptions {
                Dbs, err := GetallSQLDBs(subsID)
                if err != nil {
                        t.Errorf("Failed to get databases for subscription: %s because %v", key, err)
                }
                t.Logf("Found %d databases in %s", len(Dbs), key)
        }
}