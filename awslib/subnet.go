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
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// GetAllSubnetInstances returns a complete list of instances for a given session 
// Needs to get updated with the latest aws-sdk-go version
func GetAllSubnetInstances(sess *session.Session) ([]*ec2.Subnet, error) {
	ec2c := ec2.New(sess)
	var allInstances []*ec2.Subnet
	input := ec2.DescribeSubnetsInput{}
	result, err := ec2c.DescribeSubnets(&input)
	if err != nil {
		return allInstances, err
	}
	allInstances = result.Subnets	
	return allInstances, nil
}
