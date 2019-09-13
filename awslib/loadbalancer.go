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
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/jpillora/backoff"
)

// GetAllCLB resturns a complete list of Classic Load Balancers for a given session
func GetAllCLB(sess *session.Session) ([]*elb.LoadBalancerDescription, error) {
	lb := elb.New(sess)
	allLoadBalancersDone := false

	var allLoadBalancers []*elb.LoadBalancerDescription
	input := elb.DescribeLoadBalancersInput{}

	b := &backoff.Backoff{
		//These are the defaults
		Min:    10 * time.Millisecond,
		Max:    1 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	for !allLoadBalancersDone {
		// Describe instances with no filters
		result, err := lb.DescribeLoadBalancers(&input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				// Retry with backoff incase Rate has been exceeded
				if aerr.Code() == "RateExceeded" {
					time.Sleep(b.Duration())
					continue
				}
				return allLoadBalancers, aerr
			}
			return allLoadBalancers, err
		}
		if err != nil {
			return nil, errors.New("Error Describing LoadBalancers")
		}
		allLoadBalancers = append(allLoadBalancers, result.LoadBalancerDescriptions...)
		if result.NextMarker == nil {
			allLoadBalancersDone = true
			continue
		}
		input.SetMarker(*result.NextMarker)
	}
	return allLoadBalancers, nil
}

// GetAllALBAndNLB resturns a complete list of Application & Network Load Balancers for a given session
func GetAllALBAndNLB(sess *session.Session) ([]*elbv2.LoadBalancer, error) {
	lb := elbv2.New(sess)
	allLoadBalancersDone := false

	var allLoadBalancers []*elbv2.LoadBalancer
	input := elbv2.DescribeLoadBalancersInput{}

	b := &backoff.Backoff{
		//These are the defaults
		Min:    10 * time.Millisecond,
		Max:    1 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	for !allLoadBalancersDone {
		// Describe instances with no filters
		result, err := lb.DescribeLoadBalancers(&input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				// Retry with backoff incase Rate has been exceeded
				if aerr.Code() == "RateExceeded" {
					time.Sleep(b.Duration())
					continue
				}
				return allLoadBalancers, aerr
			}
			return allLoadBalancers, err
		}
		if err != nil {
			return nil, errors.New("Error Describing LoadBalancers")
		}
		allLoadBalancers = append(allLoadBalancers, result.LoadBalancers...)
		if result.NextMarker == nil {
			allLoadBalancersDone = true
			continue
		}
		input.SetMarker(*result.NextMarker)
	}
	return allLoadBalancers, nil
}
