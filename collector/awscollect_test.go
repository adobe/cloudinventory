package collector

import (
	"github.com/tchaudhry91/cloudinventory/awslib"
	"testing"
)

// TestAWSCollectorCreation attempts to build a new collector with initialized sessions for the given partition. This test is also very credential dependent.
func TestAWSCollectorCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	// Currently only testing with default partition credentials
	for _, testCase := range []struct {
		partition string
		err       bool
	}{
		{partition: "default", err: false},
		{partition: "china", err: false},
		{partition: "non-existent", err: true},
	} {
		_, err := NewAWSCollector(testCase.partition)
		if have := (err != nil); testCase.err != have {
			t.Errorf("%s\tWant:%t\tHave:%t", testCase.partition, testCase.err, have)
		}
	}
}

// TestCollectEC2 tries to gather instances across all regions
func TestCollectEC2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	col, err := NewAWSCollector("default")
	if err != nil {
		t.Errorf("Failed to create default collector: %v", err)
	}
	ii, err := col.CollectEC2()
	if err != nil {
		// Improve this test, right now, does not test anything of note
		t.Errorf("Failed to collect EC2 instances: %v", err)
	}
	// Depending on the Account, the map should contain one of the following regions
	if len(ii) != 0 {
		for r := range ii {
			if !stringInSlice(r, awslib.GetAllRegions()) {
				t.Errorf("Found rougue region in instances")
			}
		}
	}
}

// TestCollectRDS tries to gather DB instances across all regions
func TestCollectRDS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	col, err := NewAWSCollector("default")
	if err != nil {
		t.Errorf("Failed to create default collector: %v", err)
	}
	ii, err := col.CollectRDS()
	if err != nil {
		// Improve this test, right now, does not test anything of note
		t.Errorf("Failed to collect RDS instances: %v", err)
	}
	// Depending on the Account, the map should contain one of the following regions
	if len(ii) != 0 {
		for r := range ii {
			if !stringInSlice(r, awslib.GetAllRegions()) {
				t.Errorf("Found rougue region in instances")
			}
		}
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
