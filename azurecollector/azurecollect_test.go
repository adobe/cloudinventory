package azurecollector

import "testing"

// TestCollectSQLDBs test the function CollectSQLDBs
func TestCollectSQLDBs(t *testing.T) {
        if testing.Short() {
                t.Skip("Skipping test in short mode")
        }
        col, err := NewAzureCollector()
        if err != nil {
                t.Errorf("Failed to create collector: %v", err)
        }
        maxGoRoutines := len(col.SubscriptionMap)
        _, err = col.CollectSQLDBs(maxGoRoutines)
        if err != nil {
                t.Errorf("Failed to collect SQL Databases: %v", err)
        }
}

// TestCollectVMS test the function CollectVMS
func TestCollectVMS(t *testing.T) {
        if testing.Short() {
                t.Skip("Skipping test in short mode")
        }
        col, err := NewAzureCollector()
        if err != nil {
                t.Errorf("Failed to create collector: %v", err)
        }
        maxGoRoutines := len(col.SubscriptionMap)
        _, err = col.CollectVMS(maxGoRoutines)
        if err != nil {
                t.Errorf("Failed to collect Virtual Machines: %v", err)
        }
}

// TestCollectLoadBalancers test the function CollectLoadBalancers
func TestCollectLoadBalancers(t *testing.T) {
        if testing.Short() {
                t.Skip("Skipping test in short mode")
        }
        col, err := NewAzureCollector()
        if err != nil {
                t.Errorf("Failed to create collector: %v", err)
        }
        maxGoRoutines := len(col.SubscriptionMap)
        _, _, err = col.CollectLoadBalancers(maxGoRoutines)
        if err != nil {
                t.Errorf("Failed to collect Load Balancers: %v", err)
        }
}

// TestCollectCDN test the function CollectCDN
func TestCollectCDN(t *testing.T) {
        if testing.Short() {
                t.Skip("Skipping test in short mode")
        }
        col, err := NewAzureCollector()
        if err != nil {
                t.Errorf("Failed to create collector: %v", err)
        }
        maxGoRoutines := len(col.SubscriptionMap)
        _, _, err = col.CollectCDN(maxGoRoutines)
        if err != nil {
                t.Errorf("Failed to collect CDN: %v", err)
        }
}
