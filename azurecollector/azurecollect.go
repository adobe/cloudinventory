package azurecollector

import (
        "Azure/cloudinventory/azurelib"
        "context"
        "fmt"
        "github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/v3.0/sql"
        "sync"
)

type AzureCollector struct {
        Subscriptionmap map[string]string
}

func (col *AzureCollector) InitSubscription(ctx context.Context) error {
        subscription := make(map[string]string)
        var err error
        subscription, err = azurelib.GetallSubscriptionIDs(ctx)
        if err != nil {
                return err
        }
        col.Subscriptionmap = subscription
        return nil

}

func (col AzureCollector) CollectSQLDBs() (map[string][]*sql.Database, error) {
        DBs := make(map[string][]*sql.Database)
        type DBspersubscriptionID struct {
                subscriptionName string
                dblist           []*sql.Database
        }
        dbsChan := make(chan DBspersubscriptionID, len(col.Subscriptionmap))
        errChan := make(chan error, len(col.Subscriptionmap))

        var wg sync.WaitGroup

        for subscriptionName, subID := range col.Subscriptionmap {
                wg.Add(1)
                go func(subID string, subscriptionName string, dbsChan chan DBspersubscriptionID, errChan chan error) {
                        defer wg.Done()
                        dbs, err := CollectSQLDBspersubscriptionID(subID)
                        if err != nil {
                                errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", subscriptionName, err))
                                return
                        }
                        if dbs == nil {
                                return
                        }
                        dbsChan <- DBspersubscriptionID{subscriptionName, dbs}
                }(subID, subscriptionName, dbsChan, errChan)
        }
        wg.Wait()
        close(dbsChan)
        close(errChan)

        if len(errChan) > 0 {
                return nil, fmt.Errorf(fmt.Sprintf("Failed to gather SQL databases Data: %v", <-errChan))
        }

        for subscriptionDbs := range dbsChan {
                DBs[subscriptionDbs.subscriptionName] = subscriptionDbs.dblist
        }

        return DBs, nil

}

func CollectSQLDBspersubscriptionID(subscriptionID string) ([]*sql.Database, error) {

        dblist, err := azurelib.GetallSQLDBs(subscriptionID)
        return dblist, err
}