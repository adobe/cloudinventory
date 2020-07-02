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
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jpillora/backoff"
)

// GetAllSubnetInstances returns a complete list of instances for a given session
func GetAllSubnetInstances(sess *session.Session) ([]*ec2.Subnet, error) {
	ec2c := ec2.New(sess)
	allInstancesDone := false
	var allInstances []*ec2.Subnet
	input := ec2.DescribeSubnetsInput{}
	b := &backoff.Backoff{
		//These are the defaults
		Min:    10 * time.Millisecond,
		Max:    30 * time.Second,
		Factor: 2,
		Jitter: false,
	}
	for !allInstancesDone {
		// Describe instances with no filters
		result, err := ec2c.DescribeSubnets(&input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				// Retry with backoff if RateExceeded
				if aerr.Code() == "RateExceeded" {
					time.Sleep(b.Duration())
					continue
				}

				return allInstances, aerr

			}

			return allInstances, err
		}
		b.Reset()
		for _, reservation := range result.Subnets  {
			allInstances = append(allInstances, reservation)
		}
		if result.NextToken == nil {
			allInstancesDone = true
			continue
		}
		input.SetNextToken(*result.NextToken)
	}
	return allInstances, nil
}
