package azurelib

import (
        "context"
        "github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/v3.0/sql"
        "github.com/Azure/go-autorest/autorest/azure/auth"
        "strings"
        "sync"
        "time"
)

type additionalInfo struct {
        SubscriptionID string
        ResourceGroup  string
        ServerName     string
}
type SQLDBInfo struct {
        BasicInfo      *sql.Database
        AdditionalInfo additionalInfo
}

// GetAllSQLDBs function returns list of SQL databases for a given subscriptionID
func GetAllSQLDBs(subscriptionID string) (DBList []*SQLDBInfo, err error) {
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
                results, errs := dataClient.ListByServerComplete(ctx, resourceGroup, serverName)
                err = errs
                if err != nil {
                        return
                }
                instancesChan := make(chan *SQLDBInfo, 800)
                var wg sync.WaitGroup
                for results.NotDone() {
                        wg.Add(1)
                        var dbInfo SQLDBInfo
                        var addInfo additionalInfo
                        addInfo.ResourceGroup = resourceGroup
                        addInfo.ServerName = serverName
                        addInfo.SubscriptionID = subscriptionID
                        dbInfo.AdditionalInfo = addInfo
                        db := results.Value()
                        dbInfo.BasicInfo = &db
                        go func(instancesChan chan *SQLDBInfo) {
                                defer wg.Done()
                                instancesChan <- &dbInfo
                        }(instancesChan)
                        if err = results.Next(); err != nil {
                                return
                        }
                }
                wg.Wait()
                close(instancesChan)
                for Db := range instancesChan {
                        DBList = append(DBList, Db)
                }
                if err = server.Next(); err != nil {
                        return
                }
        }
        return
}
