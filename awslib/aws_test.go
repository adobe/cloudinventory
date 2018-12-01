package awslib

import (
	"testing"
)

// TestGetAllInstances checks if the lib is able to gather all instances properly or not.
// This test REQUIRES a working AWS account and credentials to read from EC2
// Since the validity of the test depends on the account itself, this test is written to be as generic as possible
// This test does NOT fail unless there is an error in the gathering, the gathering itself is not validated.TestGetAllInstances
// Use --verbose to see instance numbers
func TestGetAllInstances(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	sessions, err := BuildSessions(GetAllRegions())
	if err != nil {
		t.Errorf("Unable to get sessions: %v", err)
	}
	for r, sess := range sessions {
		instances, err := GetAllInstances(sess)
		if err != nil {
			t.Errorf("Failed to get Instances for region: %s because %v", r, err)
		}
		t.Logf("Found %d instances in %s", len(instances), r)
	}
}
