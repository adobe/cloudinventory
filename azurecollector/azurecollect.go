package azurecollector

import (
        "context"
        "fmt"
        "github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/v3.0/sql"
        "github.com/Thushara67/cloudInventoryforAzure/azurelib"
        "sync"
        "time"
)

// AzureCollector is a struct that contains a map of subscription name and its subscriptionID
type AzureCollector struct {
        SubscriptionMap map[string]string
}

// NewAzureCollector returns an AzureCollector with subscription info set in subscriptionMap.
func NewAzureCollector() (AzureCollector, error) {
        var col AzureCollector
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
        err := col.GetSubscription(ctx)
        if err != nil {
                return col, err
        }
        return col, nil
}

// GetSubscription adds map of subscription name with subscriptionID to the AzureCollector
func (col *AzureCollector) GetSubscription(ctx context.Context) error {
        subscription := make(map[string]string)
        var err error
        subscription, err = azurelib.GetAllSubscriptionIDs(ctx)
        if err != nil {
                return err
        }
        col.SubscriptionMap = subscription
        return nil

}

// CollectVMS gathers all the Virtual Machines for each subscriptionID in an account level
func (col *AzureCollector) CollectVMS() (map[string][]*azurelib.VirtualMachineInfo, error) {
        subscriptionsMap := make(map[string][]*azurelib.VirtualMachineInfo)
        type subscriptionsVMS struct {
                subscriptionName string
                VMList           []*azurelib.VirtualMachineInfo
        }

        subscriptionsChan := make(chan subscriptionsVMS, len(col.SubscriptionMap))
        errChan := make(chan error, len(col.SubscriptionMap))
        var wg sync.WaitGroup

        for subscriptionName, subscriptionID := range col.SubscriptionMap {
                wg.Add(1)
                go func(subscriptionName string, subscriptionID string, subscriptionsChan chan subscriptionsVMS, errChan chan error) {
                        defer wg.Done()

                        VMList, err := CollectVMsPerSubscriptionID(subscriptionID)
                        if err != nil {
                                errChan <- err
                                return
                        }
                        subscriptionsChan <- subscriptionsVMS{subscriptionName, VMList}
                }(subscriptionName, subscriptionID, subscriptionsChan, errChan)
        }

        wg.Wait()
        close(subscriptionsChan)
        close(errChan)

        if len(errChan) > 0 {
                return nil, fmt.Errorf(fmt.Sprintf("Failed to gather VM Data: %v", <-errChan))
        }

        for subsVMS := range subscriptionsChan {
                subscriptionsMap[subsVMS.subscriptionName] = subsVMS.VMList
        }

        return subscriptionsMap, nil
}

// CollectSQLDBs gathers SQL databases for each subscriptionID in an account level
func (col *AzureCollector) CollectSQLDBs() (map[string][]*sql.Database, error) {
        DBs := make(map[string][]*sql.Database)
        type DBsPerSubscriptionID struct {
                subscriptionName string
                dbList           []*sql.Database
        }
        dbsChan := make(chan DBsPerSubscriptionID, len(col.SubscriptionMap))
        errChan := make(chan error, len(col.SubscriptionMap))

        var wg sync.WaitGroup

        for subscriptionName, subID := range col.SubscriptionMap {
                wg.Add(1)
                go func(subID string, subscriptionName string, dbsChan chan DBsPerSubscriptionID, errChan chan error) {
                        defer wg.Done()
                        dbs, err := CollectSQLDBsPerSubscriptionID(subID)
                        if err != nil {
                                errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", subscriptionName, err))
                                return
                        }
                        if dbs == nil {
                                return
                        }
                        dbsChan <- DBsPerSubscriptionID{subscriptionName, dbs}
                }(subID, subscriptionName, dbsChan, errChan)
        }
        wg.Wait()
        close(dbsChan)
        close(errChan)

        if len(errChan) > 0 {
                return nil, fmt.Errorf(fmt.Sprintf("Failed to gather SQL databases Data: %v", <-errChan))
        }

        for subscriptionDbs := range dbsChan {
                DBs[subscriptionDbs.subscriptionName] = subscriptionDbs.dbList
        }

        return DBs, nil

}

// CollectSQLDBsPerSubscriptionID returns a slice of SQL databases for a given subscriptionID
func CollectSQLDBsPerSubscriptionID(subscriptionID string) ([]*sql.Database, error) {

        dblist, err := azurelib.GetAllSQLDBs(subscriptionID)
        return dblist, err
}

// CollectVMsPerSubscriptionID returns a slice of VirtualMachineInfo for a given subscriptionID
func CollectVMsPerSubscriptionID(subscriptionID string) ([]*azurelib.VirtualMachineInfo, error) {

        var vmList []*azurelib.VirtualMachineInfo
        client := azurelib.GetNewClients(subscriptionID)
        err := client.AuthorizeClients()
        if err != nil {
                return vmList, err
        }
        vmList, err = azurelib.GetAllVMS(client)
        return vmList, err
}
