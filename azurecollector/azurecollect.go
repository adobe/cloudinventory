package azurecollector

import (
        "context"
        "fmt"
        "github.com/adobe/cloudinventory/azurelib"
        "github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/network/mgmt/network"
	"github.com/Azure/azure-sdk-for-go/services/cdn/mgmt/2019-04-15/cdn"
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
                subID["subscription_id  : "+subscriptionID[i]] = subscriptionID[i]
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
// Function takes no.of goroutines to be created from user as input
func (col *AzureCollector) CollectVMS(maxGoRoutines int) (map[string][]*azurelib.VirtualMachineInfo, error) {
        subscriptionsMap := make(map[string][]*azurelib.VirtualMachineInfo)
        type subscriptionsVMS struct {
                subscriptionName string
                VMList           []*azurelib.VirtualMachineInfo
        }
        var chanCapacity int

        if maxGoRoutines >= len(col.SubscriptionMap) || maxGoRoutines < 0 {
                chanCapacity = len(col.SubscriptionMap)
        } else {
                chanCapacity = maxGoRoutines
        }
        subscriptionCount := 0
        subscriptionsChan := make(chan subscriptionsVMS, chanCapacity)
        errChan := make(chan error, chanCapacity)

        var wg sync.WaitGroup

        for subscriptionName, subscriptionID := range col.SubscriptionMap {
                if subscriptionCount < chanCapacity {
                        wg.Add(1)
                        go func(subscriptionName string, subscriptionID string, subscriptionsChan chan subscriptionsVMS, errChan chan error) {
                                defer wg.Done()
                                VMList, err := CollectVMsPerSubscriptionID(subscriptionID)
                                if err != nil {
                                        errChan <- err
                                        return
                                }
                                // Ignore subscriptions with no instances
                                if VMList == nil {
                                        return
                                }
                                subscriptionsChan <- subscriptionsVMS{subscriptionName, VMList}
                        }(subscriptionName, subscriptionID, subscriptionsChan, errChan)
                        if subscriptionCount == chanCapacity-1 {
                                wg.Wait()
                                close(subscriptionsChan)
                                close(errChan)
                                if len(errChan) > 0 {
                                        return nil, fmt.Errorf(fmt.Sprintf("Failed to gather VM Data: %v", <-errChan))
                                }
                                for subsVMS := range subscriptionsChan {
                                        subscriptionsMap[subsVMS.subscriptionName] = subsVMS.VMList
                                }
                        }
                } else {
                        VMList, err := CollectVMsPerSubscriptionID(subscriptionID)
                        if err != nil {
                                return nil, fmt.Errorf(fmt.Sprintf("Failed to gather VM Data: %v", err))
                        }
                        // Ignore subscriptions with no instances
                        if VMList == nil {
                                subscriptionCount++
                                continue
                        }
                        subscriptionsMap[subscriptionName] = VMList
                }
                subscriptionCount++
        }
        return subscriptionsMap, nil
}

// CollectSQLDBs gathers SQL databases for each subscriptionID in an account level
// Function takes no.of goroutines to be created from user as input
func (col *AzureCollector) CollectSQLDBs(maxGoRoutines int) (map[string][]*azurelib.SQLDBInfo, error) {
        DBs := make(map[string][]*azurelib.SQLDBInfo)
        type DBsPerSubscriptionID struct {
                subscriptionName string
                dbList           []*azurelib.SQLDBInfo
        }
        var chanCapacity int

        if maxGoRoutines >= len(col.SubscriptionMap) || maxGoRoutines < 0 {
                chanCapacity = len(col.SubscriptionMap)
        } else {
                chanCapacity = maxGoRoutines
        }
        subscriptionCount := 0
        dbsChan := make(chan DBsPerSubscriptionID, chanCapacity)
        errChan := make(chan error, chanCapacity)

        var wg sync.WaitGroup

        for subscriptionName, subID := range col.SubscriptionMap {
                if subscriptionCount < chanCapacity {
                        wg.Add(1)
                        go func(subID string, subscriptionName string, dbsChan chan DBsPerSubscriptionID, errChan chan error) {
                                defer wg.Done()
                                dbs, err := CollectSQLDBsPerSubscriptionID(subID)
                                if err != nil {
                                        errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", subscriptionName, err))
                                        return
                                }
                                // Ignore subscriptions with no sql database instances
                                if dbs == nil {
                                        return
                                }
                                dbsChan <- DBsPerSubscriptionID{subscriptionName, dbs}
                        }(subID, subscriptionName, dbsChan, errChan)
                        if subscriptionCount == chanCapacity-1 {
                                wg.Wait()
                                close(dbsChan)
                                close(errChan)

                                if len(errChan) > 0 {
                                        return nil, fmt.Errorf(fmt.Sprintf("Failed to gather SQL databases Data: %v", <-errChan))
                                }

                                for subscriptionDbs := range dbsChan {
                                        DBs[subscriptionDbs.subscriptionName] = subscriptionDbs.dbList
                                }
                        }
                } else {
                        dbs, err := CollectSQLDBsPerSubscriptionID(subID)
                        if err != nil {
                                return nil, fmt.Errorf(fmt.Sprintf("Failed to gather SQL databases Data: %v", err))
                        }
                        // Ignore subscriptions with no  sql database instances
                        if dbs == nil {
                                subscriptionCount++
                                continue
                        }
                        DBs[subscriptionName] = dbs
                }
                subscriptionCount++
        }

        return DBs, nil

}

// CollectCDN gathers information and stats of CDN for each subscriptionID in an account level
// Function takes no.of goroutines to be created from user as input
func (col *AzureCollector) CollectCDN(maxGoRoutines int) (map[string][]*cdn.Endpoint, map[string]int, error) {
        subscriptionsMap := make(map[string][]*cdn.Endpoint)
        CDNCount := make(map[string]int)
        type subscriptionsCDN struct {
                subscriptionName string
                CDNList           []*cdn.Endpoint
        }
        var chanCapacity int

        if maxGoRoutines >= len(col.SubscriptionMap) || maxGoRoutines < 0 {
                chanCapacity = len(col.SubscriptionMap)
        } else {
                chanCapacity = maxGoRoutines
        }
        subscriptionCount := 0
        subscriptionsChan := make(chan subscriptionsCDN, chanCapacity)
        errChan := make(chan error, chanCapacity)

        var wg sync.WaitGroup

        for subscriptionName, subscriptionID := range col.SubscriptionMap {
                if subscriptionCount < chanCapacity {
                        wg.Add(1)
                        go func(subscriptionName string, subscriptionID string, subscriptionsChan chan subscriptionsCDN, errChan chan error) {
                                defer wg.Done()
                                CDNList, err := CollectCDNPerSubscriptionID(subscriptionID)
                                if err != nil {
                                        errChan <- err
                                        return
                                }
                                // Ignore subscriptions with no instances
                                if CDNList == nil {
                                        return
                                }
                                subscriptionsChan <- subscriptionsCDN{subscriptionName, CDNList}
                        }(subscriptionName, subscriptionID, subscriptionsChan, errChan)
                        if subscriptionCount == chanCapacity-1 {
                                wg.Wait()
                                close(subscriptionsChan)
                                close(errChan)
                                if len(errChan) > 0 {
                                        return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather CDN Data: %v", <-errChan))
                                }
                                for subsCDN := range subscriptionsChan {
                                        subscriptionsMap[subsCDN.subscriptionName] = subsCDN.CDNList
                                        CDNCount[subsCDN.subscriptionName] = len(subsCDN.CDNList)
                                }
                        }
                } else {
                        CDNList, err := CollectCDNPerSubscriptionID(subscriptionID)
                        if err != nil {
                                return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather CDN Data: %v", err))
                        }
                        // Ignore subscriptions with no instances
                        if CDNList == nil {
                                subscriptionCount++
                                continue
                        }
                        subscriptionsMap[subscriptionName] = CDNList
                        CDNCount[subscriptionName] =len(CDNList)
                }
                subscriptionCount++
        }
        return subscriptionsMap, CDNCount, nil
}

// CollectLoadBalancers gathers Load Balancers stats and data for each subscriptionID in an account level
// Function takes no.of goroutines to be created from user as input
func (col *AzureCollector) CollectLoadBalancers(maxGoRoutines int) (map[string][]*network.LoadBalancer, map[string]int, error) {
        LDBs := make(map[string][]*network.LoadBalancer)
        LDBCount := make(map[string]int)
        type LoadBalancersPerSubscriptionID struct {
                SubscriptionName string
                LdbList          []*network.LoadBalancer
        }
        var chanCapacity int

        if maxGoRoutines >= len(col.SubscriptionMap) || maxGoRoutines < 0 {
                chanCapacity = len(col.SubscriptionMap)
        } else {
                chanCapacity = maxGoRoutines
        }
        subscriptionCount := 0
        ldbsChan := make(chan LoadBalancersPerSubscriptionID, chanCapacity)
        errChan := make(chan error, chanCapacity)

        var wg sync.WaitGroup

        for subscriptionName, subID := range col.SubscriptionMap {
                if subscriptionCount < chanCapacity {
                        wg.Add(1)
                        go func(subID string, subscriptionName string, ldbsChan chan LoadBalancersPerSubscriptionID, errChan chan error) {
                                defer wg.Done()
                                ldbs, err := CollectLoadBalancersPerSubscriptionID(subID)
                                if err != nil {
                                        errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", subscriptionName, err))
                                        return
                                }
                                // Ignore subscriptions with no load balancer instances
                                if ldbs == nil {
                                        return
                                }
                                ldbsChan <- LoadBalancersPerSubscriptionID{subscriptionName, ldbs}
                        }(subID, subscriptionName, ldbsChan, errChan)
                        if subscriptionCount == chanCapacity-1 {
                                wg.Wait()
                                close(ldbsChan)
                                close(errChan)
                                if len(errChan) > 0 {
                                        return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather load balancers Data: %v", <-errChan))
                                }
                                for subscriptionLDBs := range ldbsChan {
                                        LDBs[subscriptionLDBs.SubscriptionName] = subscriptionLDBs.LdbList
                                        LDBCount[subscriptionLDBs.SubscriptionName] = len(subscriptionLDBs.LdbList)
                                }
                        }
                } else {
                        ldbs, err := CollectLoadBalancersPerSubscriptionID(subID)
                        if err != nil {
                                return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather load balancers Data: %v", err))
                        }
                        // Ignore subscriptions with no load balancer instances
                        if ldbs == nil {
                                subscriptionCount++
                                continue
                        }
                        LDBs[subscriptionName] = ldbs
                        LDBCount[subscriptionName] = len(ldbs)
                }
                subscriptionCount++
        }
        return LDBs, LDBCount, nil
}

// CollectVMSCount gathers virtual machine stats for each subscriptionID in an account level
// Function takes no.of goroutines to be created from user as input
func (col *AzureCollector) CollectVMSCount(maxGoRoutines int) (map[string]int, error) {
        subscriptionsStat := make(map[string]int)
        type VMCountPerSubscriptionID struct {
                SubscriptionName string
                VMCount          int
        }
        var chanCapacity int

        if maxGoRoutines >= len(col.SubscriptionMap) || maxGoRoutines < 0 {
                chanCapacity = len(col.SubscriptionMap)
        } else {
                chanCapacity = maxGoRoutines
        }
        subscriptionCount := 0
        subscriptionsChan := make(chan VMCountPerSubscriptionID, chanCapacity)
        errChan := make(chan error, chanCapacity)

        var wg sync.WaitGroup

        for subscriptionName, subscriptionID := range col.SubscriptionMap {
                if subscriptionCount < chanCapacity {
                        wg.Add(1)
                        go func(subscriptionName string, subscriptionID string, subscriptionsChan chan VMCountPerSubscriptionID, errChan chan error) {
                                defer wg.Done()
                                vmCount, err := azurelib.GetVMCount(subscriptionID)
                                if err != nil {
                                        errChan <- err
                                        return
                                }
                                // Ignore subscriptions with no instances
                                if vmCount == 0 {
                                        return
                                }
                                subscriptionsChan <- VMCountPerSubscriptionID{subscriptionName, vmCount}
                        }(subscriptionName, subscriptionID, subscriptionsChan, errChan)
                        if subscriptionCount == chanCapacity-1 {
                                wg.Wait()
                                close(subscriptionsChan)
                                close(errChan)
                                if len(errChan) > 0 {
                                        return nil, fmt.Errorf(fmt.Sprintf("Failed to gather VM Count: %v", <-errChan))
                                }
                                for subsVMS := range subscriptionsChan {
                                        subscriptionsStat[subsVMS.SubscriptionName] = subsVMS.VMCount
                                }
                        }
                } else {
                        vmCount, err := azurelib.GetVMCount(subscriptionID)
                        if err != nil {
                                return nil, fmt.Errorf(fmt.Sprintf("Failed to gather VM Count: %v", err))
                        }
                        // Ignore subscriptions with no instances
                        if vmCount == 0 {
                                subscriptionCount++
                                continue
                        }
                        subscriptionsStat[subscriptionName] = vmCount
                }
                subscriptionCount++
        }
        return subscriptionsStat, nil
}

// CollectSQLDBCount gathers sql databases stats for each subscriptionID in an account level
// Function takes no.of goroutines to be created from user as input
func (col *AzureCollector) CollectSQLDBCount(maxGoRoutines int) (map[string]int, error) {
        subscriptionsStat := make(map[string]int)
        type SQLDBCountPerSubscriptionID struct {
                SubscriptionName string
                SQLDBCount       int
        }

        var chanCapacity int

        if maxGoRoutines >= len(col.SubscriptionMap) || maxGoRoutines < 0 {
                chanCapacity = len(col.SubscriptionMap)
        } else {
                chanCapacity = maxGoRoutines
        }
        subscriptionCount := 0
        subscriptionsChan := make(chan SQLDBCountPerSubscriptionID, chanCapacity)
        errChan := make(chan error, chanCapacity)

        var wg sync.WaitGroup
        for subscriptionName, subscriptionID := range col.SubscriptionMap {
                if subscriptionCount < chanCapacity {
                        wg.Add(1)
                        go func(subscriptionName string, subscriptionID string, subscriptionsChan chan SQLDBCountPerSubscriptionID, errChan chan error) {
                                defer wg.Done()
                                sqldbCount, err := azurelib.GetSQLDBCount(subscriptionID)
                                if err != nil {
                                        errChan <- err
                                        return
                                }
                                // Ignore subscriptions with no sqldatabase instances
                                if sqldbCount == 0 {
                                        return
                                }
                                subscriptionsChan <- SQLDBCountPerSubscriptionID{subscriptionName, sqldbCount}
                        }(subscriptionName, subscriptionID, subscriptionsChan, errChan)
                        if subscriptionCount == chanCapacity-1 {
                                wg.Wait()
                                close(subscriptionsChan)
                                close(errChan)
                                if len(errChan) > 0 {
                                        return nil, fmt.Errorf(fmt.Sprintf("Failed to gather SQLDB Count: %v", <-errChan))
                                }
                                for subsVMS := range subscriptionsChan {
                                        subscriptionsStat[subsVMS.SubscriptionName] = subsVMS.SQLDBCount
                                }
                        }
                } else {
                        sqldbCount, err := azurelib.GetSQLDBCount(subscriptionID)
                        if err != nil {
                                return nil, fmt.Errorf(fmt.Sprintf("Failed to gather SQLDB Count: %v", err))
                        }
                        // Ignore subscriptions with no sql database instances
                        if sqldbCount == 0 {
                                subscriptionCount++
                                continue
                        }
                        subscriptionsStat[subscriptionName] = sqldbCount
                }
                subscriptionCount++
        }

        return subscriptionsStat, nil
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

// CollectCDNPerSubscriptionID returns a slice of CDN for a given subscriptionID
func CollectCDNPerSubscriptionID(subscriptionID string) ([]*cdn.Endpoint, error) {

        cdnList, err := azurelib.GetAllCDN(subscriptionID)
        return cdnList, err
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
