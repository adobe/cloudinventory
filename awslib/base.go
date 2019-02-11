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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
)

// GetAllRegions returns all regions for AWS except US-Gov and China
func GetAllRegions() []string {
	awsRegions := endpoints.AwsPartition().Regions()
	var regions []string
	for _, r := range awsRegions {
		regions = append(regions, r.ID())
	}
	return regions
}

// GetAllChinaRegions returns all regions for the AWS China Partition
func GetAllChinaRegions() []string {
	awsRegions := endpoints.AwsCnPartition().Regions()
	var regions []string
	for _, r := range awsRegions {
		regions = append(regions, r.ID())
	}
	return regions
}

// BuildSessions returns a map of sessions for each region using Environment Credentials
func BuildSessions(regions []string) (map[string]*session.Session, error) {
	creds := credentials.NewEnvCredentials()
	return BuildSessionsWithCredentials(regions, creds)
}

// BuildSessionsWithCredentials returns a map of sessions for each region using supplied credentials
func BuildSessionsWithCredentials(regions []string, creds *credentials.Credentials) (map[string]*session.Session, error) {
	sessions := make(map[string]*session.Session)
	var errMain error
	for _, region := range regions {
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(region),
			Credentials: creds,
		})
		if err != nil {
			errMain = err
		}
		_, err = sess.Config.Credentials.Get()
		if err != nil {
			errMain = fmt.Errorf("Failed to get AWS Credentials")
			break
		}
		sessions[region] = sess
	}
	return sessions, errMain
}
