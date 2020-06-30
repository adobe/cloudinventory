/*
Copyright 2019 Adobe. All rights reserved.
This file is licensed to you under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License. You may obtain a copy
of the License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
OF ANY KIND, either express or implied. See the License for the specific language
governing permissions and limitations under the License.
*/

package awslib

import (
	"testing"
)

// TestGetAllInstances checks if the lib is able to gather all instances properly or not.
// This test REQUIRES a working AWS account and credentials to read from EC2
// Since the validity of the test depends on the account itself, this test is written to be as generic as possible
// This test does NOT fail unless there is an error in the gathering, the gathering itself is not validated.TestGetAllInstances
// Use verbose testing to see instance numbers
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

// TestGetAllDBInstances checks if the lib is able to gather all instances properly or not.
// This test REQUIRES a working AWS account and credentials to read from RDS
// Since the validity of the test depends on the account itself, this test is written to be as generic as possible
// This test does NOT fail unless there is an error in the gathering, the gathering itself is not validated.TestGetAllInstances
// Use verbose testing to see instance numbers
func TestGetAllDBInstances(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	sessions, err := BuildSessions(GetAllRegions())
	if err != nil {
		t.Errorf("Unable to get sessions : %v", err)
	}
	for r, sess := range sessions {
		dbinstances, err := GetAllDBInstances(sess)
		if err != nil {
			t.Errorf("Failed to get Instances for region: %s because %v", r, err)
		}
		t.Logf("Found %d instances in %s", len(dbinstances), r)
	}
}

// TestGetAllCDNInstances checks if the lib is able to gather all instances properly or not.
func TestGetAllCDNInstances(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	sessions, err := BuildSessions(GetAllRegions())
	if err != nil {
		t.Errorf("Unable to get sessions : %v", err)
	}
	for r, sess := range sessions {
		cdninstances, err := GetAllCDNInstances(sess)
		if err != nil {
			t.Errorf("Failed to get Instances for region: %s because %v", r, err)
		}
		t.Logf("Found %d instances in %s", len(cdninstances), r)
	}
}

