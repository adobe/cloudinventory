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

	awsRegions := GetAllRegions(false)
	for _, region := range awsRegionSample {
		if !stringInSlice(region, awsRegions) {
			t.Errorf("Could not find region %s in retrieved list: %v", region, awsRegions)
		}
	}

	// Test the same for China
	awsRegions = GetAllRegions(true)
	for _, region := range awsChinaSample {
		if !stringInSlice(region, awsRegions) {
			t.Errorf("Could not find region %s in retrieved list: %v", region, awsRegions)
		}
	}
}
