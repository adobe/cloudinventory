package azurecollector

import (
        "context"
        "fmt"
        "github.com/adobe/cloudinventory/azurelib"
        "github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/network/mgmt/network"
        "strconv"
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

// NewAzureCollectorUserDefined returns an AzureCollector with subscription info given by user in subscriptionMap.
func NewAzureCollectorUserDefined(subscriptionID []string) (AzureCollector, error) {
        var col AzureCollector
        subID := make(map[string]string)
        for i := 0; i < len(subscriptionID); i++ {
                s := strconv.Itoa(i)
                subID["SubscriptionID "+s+" : "+subscriptionID[i]] = subscriptionID[i]
        }
        col.SubscriptionMap = subID
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
func (col *AzureCollector) CollectSQLDBs() (map[string][]*azurelib.SQLDBInfo, error) {
        DBs := make(map[string][]*azurelib.SQLDBInfo)
        type DBsPerSubscriptionID struct {
                subscriptionName string
                dbList           []*azurelib.SQLDBInfo
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

// CollectLoadBalancers gathers Load Balancers for each subscriptionID in an account level
func (col *AzureCollector) CollectLoadBalancers() (map[string][]*network.LoadBalancer, error) {
	LDBs := make(map[string][]*network.LoadBalancer)
	type LoadBalancersPerSubscriptionID struct {
			SubscriptionName string
			LdbList           []*network.LoadBalancer
	}
	ldbsChan := make(chan LoadBalancersPerSubscriptionID, len(col.SubscriptionMap))
	errChan := make(chan error, len(col.SubscriptionMap))
	var wg sync.WaitGroup
	for subscriptionName, subID := range col.SubscriptionMap {
			wg.Add(1)
			go func(subID string, subscriptionName string, ldbsChan chan LoadBalancersPerSubscriptionID, errChan chan error) {
					defer wg.Done()
					ldbs, err := CollectLoadBalancersPerSubscriptionID(subID)
					if err != nil {
							errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", subscriptionName, err))
							return
					}
					if ldbs == nil {
							return
					}
					ldbsChan <- LoadBalancersPerSubscriptionID{subscriptionName, ldbs}
			}(subID, subscriptionName, ldbsChan, errChan)
	}
	wg.Wait()
	close(ldbsChan)
	close(errChan)
	if len(errChan) > 0 {
			return nil, fmt.Errorf(fmt.Sprintf("Failed to gather load balancers Data: %v", <-errChan))
	}
	for subscriptionLDBs := range ldbsChan {
			LDBs[subscriptionLDBs.SubscriptionName] = subscriptionLDBs.LdbList
	}
	return LDBs, nil
}

// CollectLoadBalancersPerSubscriptionID returns a slice of Load Balancers for a given subscriptionID
func CollectLoadBalancersPerSubscriptionID(subscriptionID string) ([]*network.LoadBalancer, error) {
	ldblist, err := azurelib.GetAllLdb(subscriptionID)
	return ldblist, err
}

// CollectSQLDBsPerSubscriptionID returns a slice of SQL databases for a given subscriptionID
func CollectSQLDBsPerSubscriptionID(subscriptionID string) ([]*azurelib.SQLDBInfo, error) {

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
