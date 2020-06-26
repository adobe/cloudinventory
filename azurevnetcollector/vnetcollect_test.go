package azurevnetcollector

import "testing"

// TestCollectVirtualNetworks test the function CollectVirtualNetworks
func TestCollectVirtualNetworks(t *testing.T) {
        if testing.Short() {
                t.Skip("Skipping test in short mode")
        }
        col, err := NewAzureCollector()
        if err != nil {
                t.Errorf("Failed to create collector: %v", err)
        }
        _, err = col.CollectVirtualNetworks()
        if err != nil {
                t.Errorf("Failed to collect Virtual networks: %v", err)
        }
}
