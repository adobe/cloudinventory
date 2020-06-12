package azurecollector

import (
        "context"
        "fmt"
        "github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/v3.0/sql"
        "github.com/Thushara67/cloudInventoryforAzure/azurelib"
        "sync"
)

//AzureCollector is a struct that contains a map of subscription name and its subscriptionID
type AzureCollector struct {
        Subscriptionmap map[string]string
}

//InitSubscription adds map of subscription name with subscriptionID to the AzureCollector
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

//CollectVMS gathers all the Virtual Machines for each subscriptionID in an account level
func (col *AzureCollector) CollectVMS() (map[string][]*azurelib.VirtualMachineinfo, error){
        subscriptionsMap := make(map[string][]*azurelib.VirtualMachineinfo)
        type subscriptionsVMS struct{
                subscriptionName string
                VMList []*azurelib.VirtualMachineinfo
        }
        
        subscriptionsChan := make(chan subscriptionsVMS, len(col.Subscriptionmap))
        errChan := make(chan error, len(col.Subscriptionmap))
        var wg sync.WaitGroup
        
        for subscriptionName, subscriptionID := range col.Subscriptionmap {
                wg.Add(1)
                go func (subscriptionName string, subscriptionID string, subscriptionsChan chan subscriptionsVMS, errChan chan error) {
                        defer wg.Done()

                        VMList, err := CollectVMspersubscriptionID(subscriptionID)
                        if err != nil {
                                errChan <- err
                                return
                        }
                        subscriptionsChan <- subscriptionsVMS{subscriptionName, VMList}
                } (subscriptionName, subscriptionID, subscriptionsChan, errChan)
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

//CollectSQLDBs gathers SQL databases for each subscriptionID in an account level
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

//CollectSQLDBspersubscriptionID returns a slice of SQL databases for a given subscriptionID
func CollectSQLDBspersubscriptionID(subscriptionID string) ([]*sql.Database, error) {

        dblist, err := azurelib.GetallSQLDBs(subscriptionID)
        return dblist, err
}

func CollectVMspersubscriptionID(subscriptionID string) ([]*azurelib.VirtualMachineinfo, error) {

        dblist, err := azurelib.GetallVMS(subscriptionID)
        return dblist, err
}
