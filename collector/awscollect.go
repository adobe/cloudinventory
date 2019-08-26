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

package collector

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/adobe/cloudinventory/awslib"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/jpillora/backoff"
)

// NewAWSCollector returns an AWSCollector with initialized sessions.
// Uses supplied credentials, Standard Environment variables if creds not specified
func NewAWSCollector(partition string, creds *credentials.Credentials) (AWSCollector, error) {
	var col AWSCollector
	regions := col.getRegions(partition)
	if regions == nil {
		return col, fmt.Errorf("Invalid Region Selected")
	}
	err := col.initSessions(regions, creds)
	if err != nil {
		return col, err
	}
	if !col.CheckCredentials() {
		return col, fmt.Errorf("Error obtaining AWS Credentials")
	}

	return col, nil
}

// AWSCollector is a concurrent inventory collection struct for Amazon Web Services
type AWSCollector struct {
	sessions map[string]*session.Session
}

func (col *AWSCollector) getRegions(partition string) []string {
	var regions []string
	switch part := strings.ToLower(partition); part {
	case "china":
		regions = awslib.GetAllChinaRegions()
	case "default":
		regions = awslib.GetAllRegions()
	default:
		return nil
	}
	return regions
}

func (col *AWSCollector) initSessions(regions []string, creds *credentials.Credentials) error {
	var sessions map[string]*session.Session
	var err error
	if creds == nil {
		sessions, err = awslib.BuildSessions(regions)
	} else {
		sessions, err = awslib.BuildSessionsWithCredentials(regions, creds)
	}
	if err != nil {
		return fmt.Errorf("Unable to build AWS Sessions: %v", err)
	}
	col.sessions = sessions
	return nil
}

// CheckCredentials tests the proper availability of AWS Credentials in the environment
func (col AWSCollector) CheckCredentials() bool {
	for _, sess := range col.sessions {
		if sess.Config.Credentials.IsExpired() {
			return false
		}
	}
	return true
}

// CollectEC2 returns a concurrently collected EC2 inventory for all the regions
func (col AWSCollector) CollectEC2() (map[string][]*ec2.Instance, error) {
	instances := make(map[string][]*ec2.Instance)

	// instanceRegion is a struct that holds all EC2 instances in a given region
	type instanceRegion struct {
		region    string
		instances []*ec2.Instance
	}

	instancesChan := make(chan instanceRegion, len(col.sessions))
	errChan := make(chan error, len(col.sessions))
	var wg sync.WaitGroup

	for region, sess := range col.sessions {
		wg.Add(1)
		go func(sess *session.Session, region string, instancesChan chan instanceRegion, errChan chan error) {
			defer wg.Done()
			chunk, err := CollectEC2PerSession(sess)

			if err != nil {
				errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", region, err))
				return
			}

			// Ignore regions with no instances
			if chunk == nil {
				return
			}
			instancesChan <- instanceRegion{region, chunk}
		}(sess, region, instancesChan, errChan)
	}
	wg.Wait()
	close(instancesChan)
	close(errChan)

	if len(errChan) > 0 {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to gather EC2 Data: %v", <-errChan))
	}

	for regionChunk := range instancesChan {
		instances[regionChunk.region] = regionChunk.instances
	}
	return instances, nil
}

// CollectZones returns a hostedZones
func (col AWSCollector) CollectZones() ([]*route53.HostedZone, error) {

	b := &backoff.Backoff{
		//These are the defaults
		Min:    10 * time.Millisecond,
		Max:    1 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	zones := make([]*route53.HostedZone, 0)
	var nextPageExists = true
	request := &route53.ListHostedZonesInput{}

	var route53Session *session.Session

	for _, session := range col.sessions {
		route53Session = session
	}
	r53 := route53.New(route53Session)

	for nextPageExists {
		response, err := r53.ListHostedZones(request)
		if err != nil {
			time.Sleep(b.Duration())
		} else {
			for recordIndex := range response.HostedZones {
				zones = append(zones, response.HostedZones[recordIndex])
			}
			if response.IsTruncated == nil || !*response.IsTruncated {
				nextPageExists = false
				break
			}
			// Setting next page.
			request.Marker = response.NextMarker
		}
	}
	return zones, nil
}

// GetHostedZoneRecords returns the hostedzonesRecords for a particular hostedZoneId
func (col AWSCollector) GetHostedZoneRecords(hostedZoneId string) ([]*route53.ResourceRecordSet, error) {
	var nextPageExists = true

	b := &backoff.Backoff{
		//These are the defaults
		Min:    10 * time.Millisecond,
		Max:    1 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	var route53Session *session.Session

	for _, session := range col.sessions {
		route53Session = session
	}
	r53 := route53.New(route53Session)

	records := make([]*route53.ResourceRecordSet, 0)
	request := &route53.ListResourceRecordSetsInput{
		HostedZoneId: &hostedZoneId,
	}

	for nextPageExists {

		response, err := r53.ListResourceRecordSets(request)
		if err != nil {
			time.Sleep(b.Duration())
		} else {
			records = append(records, response.ResourceRecordSets...)
			if response.IsTruncated == nil || !*response.IsTruncated {
				nextPageExists = false
				break
			}
			// Setting next page.
			request.StartRecordName = response.NextRecordName
			request.StartRecordIdentifier = response.NextRecordIdentifier
			request.StartRecordType = response.NextRecordType
		}
	}
	return records, nil
}

// CollectLoadBalancers returns a concurrently collected LoadBalancers inventory for all the regions
func (col AWSCollector) CollectLoadBalancers() (map[string][]*elb.LoadBalancerDescription, error) {
	instances := make(map[string][]*elb.LoadBalancerDescription)

	// instanceRegion is a struct that holds all LoadBalancers instances in a given region
	type instanceRegion struct {
		region    string
		instances []*elb.LoadBalancerDescription
	}

	instancesChan := make(chan instanceRegion, len(col.sessions))
	errChan := make(chan error, len(col.sessions))
	var wg sync.WaitGroup

	for region, sess := range col.sessions {
		wg.Add(1)
		go func(sess *session.Session, region string, instancesChan chan instanceRegion, errChan chan error) {
			defer wg.Done()
			chunk, err := CollectLoadBalancersPerSession(sess)

			if err != nil {
				errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", region, err))
				return
			}

			// Ignore regions with no instances
			if chunk == nil {
				return
			}
			instancesChan <- instanceRegion{region, chunk}
		}(sess, region, instancesChan, errChan)
	}
	wg.Wait()
	close(instancesChan)
	close(errChan)

	if len(errChan) > 0 {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to gather LoadBalancers Data: %v", <-errChan))
	}

	for regionChunk := range instancesChan {
		instances[regionChunk.region] = regionChunk.instances
	}
	return instances, nil
}

// CollectLoadBalancersPerSession returns an LoadBalancers inventory for a given session
func CollectLoadBalancersPerSession(sess *session.Session) ([]*elb.LoadBalancerDescription, error) {
	instances, err := awslib.GetAllLoadBalancers(sess)
	return instances, err
}

// CollectRDS returns a concurrently collected RDS inventory for all the regions
func (col AWSCollector) CollectRDS() (map[string][]*rds.DBInstance, error) {
	instances := make(map[string][]*rds.DBInstance)

	// instanceRegion is a struct that holds all RDS instances in a given region
	type instanceRegion struct {
		region    string
		instances []*rds.DBInstance
	}

	instancesChan := make(chan instanceRegion, len(col.sessions))
	errChan := make(chan error, len(col.sessions))
	var wg sync.WaitGroup

	for region, sess := range col.sessions {
		wg.Add(1)
		go func(sess *session.Session, region string, instancesChan chan instanceRegion, errChan chan error) {
			defer wg.Done()
			chunk, err := CollectRDSPerSession(sess)

			if err != nil {
				errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", region, err))
				return
			}

			// Ignore regions with no instances
			if chunk == nil {
				return
			}
			instancesChan <- instanceRegion{region, chunk}
		}(sess, region, instancesChan, errChan)
	}
	wg.Wait()
	close(instancesChan)
	close(errChan)

	if len(errChan) > 0 {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to gather RDS Data: %v", <-errChan))
	}

	for regionChunk := range instancesChan {
		instances[regionChunk.region] = regionChunk.instances
	}
	return instances, nil

}

// CollectRDSPerSession returns an RDS inventory for a given session
func CollectRDSPerSession(sess *session.Session) ([]*rds.DBInstance, error) {
	instances, err := awslib.GetAllDBInstances(sess)
	return instances, err
}

// CollectEC2PerSession returns an EC2 inventory for a given session
func CollectEC2PerSession(sess *session.Session) ([]*ec2.Instance, error) {
	instances, err := awslib.GetAllInstances(sess)
	return instances, err
}

// CollectEC2PerSession returns an EC2 inventory for a given session
func CollectHostedZonePerSession(sess *session.Session) ([]*route53.HostedZone, error) {
	instances, err := awslib.GetAllHostedZones(sess)
	return instances, err
}

// CollectLoadBalancerPerSession returns an LoadBalancer inventory for a given session
func CollectLoadBalancerPerSession(sess *session.Session) ([]*route53.HostedZone, error) {
	instances, err := awslib.GetAllHostedZones(sess)
	return instances, err
}
