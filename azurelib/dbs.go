package azurelib

import (
        "context"
        "github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/v3.0/sql"
        "github.com/Azure/go-autorest/autorest/azure/auth"
        "strings"
        "time"
)

// GetallSQLDBs function returns list of SQL databases
func GetallSQLDBs(subscriptionID string) (Dblist []*sql.Database, err error) {
        authorizer, err := auth.NewAuthorizerFromEnvironment()
        if err != nil {
                return
        }
        serverClient := sql.NewServersClient(subscriptionID)
        dataClient := sql.NewDatabasesClient(subscriptionID)
        serverClient.Authorizer = authorizer
        dataClient.Authorizer = authorizer
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
        server, err := serverClient.ListComplete(ctx)
        if err != nil {
                return
        }
        for server.NotDone() {
                result := server.Value()
                ID := strings.Split(*result.ID, "/")
                resourceGroup := ID[4]
                serverName := *result.Name
                result1, err1 := dataClient.ListByServerComplete(ctx, resourceGroup, serverName)
                err = err1
                if err != nil {
                        return
                }
                for result1.NotDone() {
                        db := result1.Value()
                        Dblist = append(Dblist, &db)
                        if err = result1.Next(); err != nil {
                                return
                        }
                }
                if err = server.Next(); err != nil {
                        return
                }
        }
        return
}