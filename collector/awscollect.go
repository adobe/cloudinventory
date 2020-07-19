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
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
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

// NewAWSCollectorUserDefined function takes list of regions from user as input. 
// It returns an AWSCollector with initialized sessions.
func NewAWSCollectorUserDefined(regions []string, creds *credentials.Credentials) (AWSCollector, error) {
        var col AWSCollector
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

// CollectEC2 returns a concurrently collected EC2 inventory and stats for all the regions
// Function takes no.of goroutines to be created from user as input
func (col AWSCollector) CollectEC2(maxGoRoutines int) (map[string][]*ec2.Instance, map[string]int, error) {
        instances := make(map[string][]*ec2.Instance)
        instancesCount := make(map[string]int)

        // instanceRegion is a struct that holds all EC2 instances in a given region
        type instanceRegion struct {
                region    string
                instances []*ec2.Instance
        }

        // chanCapacity used to specify buffer channel capacity
        var chanCapacity int

        if maxGoRoutines >= len(col.sessions) || maxGoRoutines < 0 {
                chanCapacity = len(col.sessions)
        } else {
                chanCapacity = maxGoRoutines
        }

        sessionCount := 0
        instancesChan := make(chan instanceRegion, chanCapacity)
        errChan := make(chan error, chanCapacity)
        var wg sync.WaitGroup

        for region, sess := range col.sessions {
                if sessionCount < chanCapacity {
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

                        if sessionCount == chanCapacity-1 {
                                wg.Wait()
                                close(instancesChan)
                                close(errChan)

                                if len(errChan) > 0 {
                                        return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather EC2 Data: %v", <-errChan))
                                }

                                for regionChunk := range instancesChan {
                                        instances[regionChunk.region] = regionChunk.instances
                                        instancesCount[regionChunk.region] = len(regionChunk.instances)
                                }
                        }
                } else {
                        chunk, err := CollectEC2PerSession(sess)
                        if err != nil {
                                return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather EC2 Data: %v", err))
                        }
                        if chunk == nil {
                                sessionCount++
                                continue
                        }
                        instances[region] = chunk
                        instancesCount[region] = len(chunk)

                }
                sessionCount++
        }

        return instances, instancesCount, nil
}

// CollectZones returns a hostedZones and its stats
func (col AWSCollector) CollectZones() ([]*route53.HostedZone, int, error) {

        b := &backoff.Backoff{
                //These are the defaults
                Min:    10 * time.Millisecond,
                Max:    1 * time.Second,
                Factor: 2,
                Jitter: false,
        }

        zones := make([]*route53.HostedZone, 0)
        count := 0
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
        count = len(zones)
        return zones, count, nil
}

// GetHostedZoneRecords returns the hostedzonesRecords and its stats for a particular hostedZoneId
func (col AWSCollector) GetHostedZoneRecords(hostedZoneId string) ([]*route53.ResourceRecordSet, int, error) {
        var nextPageExists = true
        count := 0

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
        count = len(records)
        return records, count, nil
}

// CollectClassicLoadBalancers returns a concurrently collected LoadBalancers inventory and stats for all the
// Function takes no.of goroutines to be created from user as input
func (col AWSCollector) CollectClassicLoadBalancers(maxGoRoutines int) (map[string][]*elb.LoadBalancerDescription, map[string]int, error) {
        instances := make(map[string][]*elb.LoadBalancerDescription)
        instancesCount := make(map[string]int)

        // instanceRegion is a struct that holds all load balancers instances in a given region
        type instanceRegion struct {
                region    string
                instances []*elb.LoadBalancerDescription
        }

        // chanCapacity used to specify buffer channel capacity
        var chanCapacity int

        if maxGoRoutines >= len(col.sessions) || maxGoRoutines < 0 {
                chanCapacity = len(col.sessions)
        } else {
                chanCapacity = maxGoRoutines
        }

        sessionCount := 0
        instancesChan := make(chan instanceRegion, chanCapacity)
        errChan := make(chan error, chanCapacity)
        var wg sync.WaitGroup

        for region, sess := range col.sessions {
                if sessionCount < chanCapacity {
                        wg.Add(1)
                        go func(sess *session.Session, region string, instancesChan chan instanceRegion, errChan chan error) {
                                defer wg.Done()
                                chunk, err := CollectClassicLoadBalancerPerSession(sess)

                                if err != nil {
                                        errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", region, err))
                                        return
                                }

                                // Ignore regions with no load balancer instances
                                if chunk == nil {
                                        return
                                }
                                instancesChan <- instanceRegion{region, chunk}
                        }(sess, region, instancesChan, errChan)

                        if sessionCount == chanCapacity-1 {
                                wg.Wait()
                                close(instancesChan)
                                close(errChan)

                                if len(errChan) > 0 {
                                        return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather LoadBalancers Data: %v", <-errChan))
                                }

                                for regionChunk := range instancesChan {
                                        instances[regionChunk.region] = regionChunk.instances
                                        instancesCount[regionChunk.region] = len(regionChunk.instances)
                                }
                        }
                } else {
                        chunk, err := CollectClassicLoadBalancerPerSession(sess)
                        if err != nil {
                                return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather LoadBalancers Data: %v", err))
                        }
                        if chunk == nil {
                                sessionCount++
                                continue
                        }
                        instances[region] = chunk
                        instancesCount[region] = len(chunk)

                }
                sessionCount++
        }
        return instances, instancesCount, nil
}

// CollectApplicationAndNetworkLoadBalancers returns a concurrently collected LoadBalancers inventory and stats for all the regions
// Function takes no.of goroutines to be created from user as input
func (col AWSCollector) CollectApplicationAndNetworkLoadBalancers(maxGoRoutines int) (map[string][]*elbv2.LoadBalancer, map[string]int, error) {
        instances := make(map[string][]*elbv2.LoadBalancer)
        instancesCount := make(map[string]int)

        // instanceRegion is a struct that holds all load balancers instances in a given region
        type instanceRegion struct {
                region    string
                instances []*elbv2.LoadBalancer
        }

        // chanCapacity used to specify buffer channel capacity
        var chanCapacity int

        if maxGoRoutines >= len(col.sessions) || maxGoRoutines < 0 {
                chanCapacity = len(col.sessions)
        } else {
                chanCapacity = maxGoRoutines
        }

        sessionCount := 0
        instancesChan := make(chan instanceRegion, chanCapacity)
        errChan := make(chan error, chanCapacity)
        var wg sync.WaitGroup

        for region, sess := range col.sessions {
                if sessionCount < chanCapacity {
                        wg.Add(1)
                        go func(sess *session.Session, region string, instancesChan chan instanceRegion, errChan chan error) {
                                defer wg.Done()
                                chunk, err := CollectApplicationNetworkLoadBalancerPerSession(sess)

                                if err != nil {
                                        errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", region, err))
                                        return
                                }

                                // Ignore regions with no load balancer instances
                                if chunk == nil {
                                        return
                                }
                                instancesChan <- instanceRegion{region, chunk}
                        }(sess, region, instancesChan, errChan)

                        if sessionCount == chanCapacity-1 {
                                wg.Wait()
                                close(instancesChan)
                                close(errChan)

                                if len(errChan) > 0 {
                                        return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather LoadBalancers Data: %v", <-errChan))
                                }

                                for regionChunk := range instancesChan {
                                        instances[regionChunk.region] = regionChunk.instances
                                        instancesCount[regionChunk.region] = len(regionChunk.instances)
                                }
                        }
                } else {
                        chunk, err := CollectApplicationNetworkLoadBalancerPerSession(sess)
                        if err != nil {
                                return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather LoadBalancers Data: %v", err))
                        }
                        if chunk == nil {
                                sessionCount++
                                continue
                        }
                        instances[region] = chunk
                        instancesCount[region] = len(chunk)

                }
                sessionCount++
        }
        return instances, instancesCount, nil
}

// CollectVPC returns a concurrently collected Vpc inventory  and stats for all the regions
// Function takes no.of goroutines to be created from user as input
func (col AWSCollector) CollectVPC(maxGoRoutines int) (map[string][]*ec2.Vpc, map[string]int, error) {
        instances := make(map[string][]*ec2.Vpc)
        instancesCount := make(map[string]int)

        // instanceRegion is a struct that holds all Vpc instances in a given region
        type instanceRegion struct {
                region    string
                instances []*ec2.Vpc
        }

        // chanCapacity used to specify buffer channel capacity
        var chanCapacity int

        if maxGoRoutines >= len(col.sessions) || maxGoRoutines < 0 {
                chanCapacity = len(col.sessions)
        } else {
                chanCapacity = maxGoRoutines
        }

        sessionCount := 0
        instancesChan := make(chan instanceRegion, chanCapacity)
        errChan := make(chan error, chanCapacity)
        var wg sync.WaitGroup

        for region, sess := range col.sessions {
                if sessionCount < chanCapacity {
                        wg.Add(1)
                        go func(sess *session.Session, region string, instancesChan chan instanceRegion, errChan chan error) {
                                defer wg.Done()
                                chunk, err := CollectVPCPerSession(sess)

                                if err != nil {
                                        errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", region, err))
                                        return
                                }

                                // Ignore regions with no vpc instances
                                if chunk == nil {
                                        return
                                }
                                instancesChan <- instanceRegion{region, chunk}
                        }(sess, region, instancesChan, errChan)

                        if sessionCount == chanCapacity-1 {
                                wg.Wait()
                                close(instancesChan)
                                close(errChan)

                                if len(errChan) > 0 {
                                        return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather VPC Data: %v", <-errChan))
                                }

                                for regionChunk := range instancesChan {
                                        instances[regionChunk.region] = regionChunk.instances
                                        instancesCount[regionChunk.region] = len(regionChunk.instances)
                                }
                        }
                } else {
                        chunk, err := CollectVPCPerSession(sess)
                        if err != nil {
                                return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather VPC Data: %v", err))
                        }
                        if chunk == nil {
                                sessionCount++
                                continue
                        }
                        instances[region] = chunk
                        instancesCount[region] = len(chunk)

                }
                sessionCount++
        }
        return instances, instancesCount, nil
}

// CollectSubnets returns a concurrently collected Subnets inventory and stats for all the regions
// Function takes no.of goroutines to be created from user as input
func (col AWSCollector) CollectSubnets(maxGoRoutines int) (map[string][]*ec2.Subnet, map[string]int, error) {
        instances := make(map[string][]*ec2.Subnet)
        instancesCount := make(map[string]int)

        // instanceRegion is a struct that holds all subnet instances in a given region
        type instanceRegion struct {
                region    string
                instances []*ec2.Subnet
        }

        // chanCapacity used to specify buffer channel capacity
        var chanCapacity int

        if maxGoRoutines >= len(col.sessions) || maxGoRoutines < 0 {
                chanCapacity = len(col.sessions)
        } else {
                chanCapacity = maxGoRoutines
        }

        sessionCount := 0
        instancesChan := make(chan instanceRegion, chanCapacity)
        errChan := make(chan error, chanCapacity)
        var wg sync.WaitGroup

        for region, sess := range col.sessions {
                if sessionCount < chanCapacity {
                        wg.Add(1)
                        go func(sess *session.Session, region string, instancesChan chan instanceRegion, errChan chan error) {
                                defer wg.Done()
                                chunk, err := CollectSubnetPerSession(sess)

                                if err != nil {
                                        errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", region, err))
                                        return
                                }

                                // Ignore regions with no subnet instances
                                if chunk == nil {
                                        return
                                }
                                instancesChan <- instanceRegion{region, chunk}
                        }(sess, region, instancesChan, errChan)

                        if sessionCount == chanCapacity-1 {
                                wg.Wait()
                                close(instancesChan)
                                close(errChan)

                                if len(errChan) > 0 {
                                        return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather Subnets Data: %v", <-errChan))
                                }

                                for regionChunk := range instancesChan {
                                        instances[regionChunk.region] = regionChunk.instances
                                        instancesCount[regionChunk.region] = len(regionChunk.instances)
                                }
                        }
                } else {
                        chunk, err := CollectSubnetPerSession(sess)
                        if err != nil {
                                return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather Subnets Data: %v", err))
                        }
                        if chunk == nil {
                                sessionCount++
                                continue
                        }
                        instances[region] = chunk
                        instancesCount[region] = len(chunk)
                }
                sessionCount++
        }
        return instances, instancesCount, nil
}

// CollectCloudFront returns a concurrently collected cloud front inventory and stats for all the regions
// Function takes no.of goroutines to be created from user as input
func (col AWSCollector) CollectCloudFront(maxGoRoutines int) (map[string][]*cloudfront.DistributionSummary, map[string]int, error) {
        instances := make(map[string][]*cloudfront.DistributionSummary)
        instancesCount := make(map[string]int)
        // instanceRegion is a struct that holds all CloudFront instances in a given region
        type instanceRegion struct {
                region    string
                instances []*cloudfront.DistributionSummary
        }

        // chanCapacity used to specify buffer channel capacity
        var chanCapacity int

        if maxGoRoutines >= len(col.sessions) || maxGoRoutines < 0 {
                chanCapacity = len(col.sessions)
        } else {
                chanCapacity = maxGoRoutines
        }

        sessionCount := 0
        instancesChan := make(chan instanceRegion, chanCapacity)
        errChan := make(chan error, chanCapacity)
        var wg sync.WaitGroup

        for region, sess := range col.sessions {
                if sessionCount < chanCapacity {
                        wg.Add(1)
                        go func(sess *session.Session, region string, instancesChan chan instanceRegion, errChan chan error) {
                                defer wg.Done()
                                chunk, err := CollectCloudFrontPerSession(sess)

                                if err != nil {
                                        errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", region, err))
                                        return
                                }

                                // Ignore regions with no cloud fronts
                                if chunk == nil {
                                        return
                                }
                                instancesChan <- instanceRegion{region, chunk}
                        }(sess, region, instancesChan, errChan)

                        if sessionCount == chanCapacity-1 {
                                wg.Wait()
                                close(instancesChan)
                                close(errChan)

                                if len(errChan) > 0 {
                                        return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather CloudFront Data: %v", <-errChan))
                                }

                                for regionChunk := range instancesChan {
                                        instances[regionChunk.region] = regionChunk.instances
                                        instancesCount[regionChunk.region] = len(regionChunk.instances)
                                }
                        }
                } else {
                        chunk, err := CollectCloudFrontPerSession(sess)
                        if err != nil {
                                return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather CloudFront Data: %v", err))
                        }
                        if chunk == nil {
                                sessionCount++
                                continue
                        }
                        instances[region] = chunk
                        instancesCount[region] = len(chunk)

                }
                sessionCount++
        }
        return instances, instancesCount, nil

}

// CollectRDS returns a concurrently collected RDS inventory and stats for all the regions
// Function takes no.of goroutines to be created from user as input
func (col AWSCollector) CollectRDS(maxGoRoutines int) (map[string][]*rds.DBInstance, map[string]int, error) {
        instances := make(map[string][]*rds.DBInstance)
        instancesCount := make(map[string]int)
        // instanceRegion is a struct that holds all RDS instances in a given region
        type instanceRegion struct {
                region    string
                instances []*rds.DBInstance
        }

        // chanCapacity used to specify buffer channel capacity
        var chanCapacity int

        if maxGoRoutines >= len(col.sessions) || maxGoRoutines < 0 {
                chanCapacity = len(col.sessions)
        } else {
                chanCapacity = maxGoRoutines
        }

        sessionCount := 0
        instancesChan := make(chan instanceRegion, chanCapacity)
        errChan := make(chan error, chanCapacity)
        var wg sync.WaitGroup

        for region, sess := range col.sessions {
                if sessionCount < chanCapacity {
                        wg.Add(1)
                        go func(sess *session.Session, region string, instancesChan chan instanceRegion, errChan chan error) {
                                defer wg.Done()
                                chunk, err := CollectRDSPerSession(sess)

                                if err != nil {
                                        errChan <- fmt.Errorf(fmt.Sprintf("Error while gathering %s: %v", region, err))
                                        return
                                }

                                // Ignore regions with no rds instances
                                if chunk == nil {
                                        return
                                }
                                instancesChan <- instanceRegion{region, chunk}
                        }(sess, region, instancesChan, errChan)

                        if sessionCount == chanCapacity-1 {
                                wg.Wait()
                                close(instancesChan)
                                close(errChan)

                                if len(errChan) > 0 {
                                        return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather RDS Data: %v", <-errChan))
                                }

                                for regionChunk := range instancesChan {
                                        instances[regionChunk.region] = regionChunk.instances
                                        instancesCount[regionChunk.region] = len(regionChunk.instances)
                                }
                        }
                } else {
                        chunk, err := CollectRDSPerSession(sess)
                        if err != nil {
                                return nil, nil, fmt.Errorf(fmt.Sprintf("Failed to gather RDS Data: %v", err))
                        }
                        if chunk == nil {
                                sessionCount++
                                continue
                        }
                        instances[region] = chunk
                        instancesCount[region] = len(chunk)

                }
                sessionCount++
        }
        return instances, instancesCount, nil

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

// CollectCloudFrontPerSession returns an CloudFront inventory for a given session
func CollectCloudFrontPerSession(sess *session.Session) ([]*cloudfront.DistributionSummary, error) {
	instances, err := awslib.GetAllCDNInstances(sess)
	return instances, err
}

// CollectVPCPerSession returns an Vpc inventory for a given session
func CollectVPCPerSession(sess *session.Session) ([]*ec2.Vpc, error) {
	instances, err := awslib.GetAllVPCInstances(sess)
	return instances, err
}

// CollectSubnetPerSession returns an Subnets inventory for a given session
func CollectSubnetPerSession(sess *session.Session) ([]*ec2.Subnet, error) {
	instances, err := awslib.GetAllSubnetInstances(sess)
	return instances, err
}

// CollectHostedZonePerSession returns an EC2 inventory for a given session
func CollectHostedZonePerSession(sess *session.Session) ([]*route53.HostedZone, error) {
	instances, err := awslib.GetAllHostedZones(sess)
	return instances, err
}

// CollectClassicLoadBalancerPerSession returns an LoadBalancer inventory for a given session
func CollectClassicLoadBalancerPerSession(sess *session.Session) ([]*elb.LoadBalancerDescription, error) {
	loadbalancers, err := awslib.GetAllCLB(sess)
	return loadbalancers, err
}

// CollectApplicationNetworkLoadBalancerPerSession returns an LoadBalancer inventory for a given session
func CollectApplicationNetworkLoadBalancerPerSession(sess *session.Session) ([]*elbv2.LoadBalancer, error) {
	loadbalancers, err := awslib.GetAllALBAndNLB(sess)
	return loadbalancers, err
}
