package awslib

import (
	"testing"
)

func stringInSlice(item string, list []string) bool {
	for _, a := range list {
		if a == item {
			return true
		}
	}
	return false
}

//TestGetAllRegions tests the availability of regions returned by GetAllRegions
func TestGetAllRegions(t *testing.T) {
	awsRegionSample := []string{"ap-southeast-1", "us-west-2", "ap-northeast-1", "eu-west-2", "eu-central-1"}
	awsChinaSample := []string{"cn-north-1", "cn-northwest-1"}

	awsRegions := GetAllRegions()
	for _, region := range awsRegionSample {
		if !stringInSlice(region, awsRegions) {
			t.Errorf("Could not find region %s in retrieved list: %v", region, awsRegions)
		}
	}

	// Test the same for China
	awsRegions = GetAllChinaRegions()
	for _, region := range awsChinaSample {
		if !stringInSlice(region, awsRegions) {
			t.Errorf("Could not find region %s in retrieved list: %v", region, awsRegions)
		}
	}
}

// TestBuildSessions tests if all the regions are presents and successfully able to build sessions
func TestBuildSessions(t *testing.T) {
	regions := GetAllRegions()
	sessions, err := BuildSessions(regions)
	if err != nil {
		t.Errorf("Could not build sessions because of: %v", err)
	}
	if len(sessions) != len(regions) {
		t.Errorf("Unequal number of regions and sessions.\nRegions=%d\tSessions=%d", len(regions), len(sessions))

	}
}
