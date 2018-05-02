package collector

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/tchaudhry91/cloudinventory/awslib"
	"sync"
)

// NewAWSCollector retuns an AWSCollector with initialized sessions
func NewAWSCollector() (AWSCollector, error) {
	var col AWSCollector
	err := col.initSessions()
	if err != nil {
		return col, err
	}
	if !col.CheckCredentials() {
		return col, errors.New("Error obtaining AWS Credentials.")
	}

	return col, nil
}

// AWSCollector is a concurrent inventory collection struct for Amazon Web Services
type AWSCollector struct {
	sessions map[string]*session.Session
}

func (col *AWSCollector) initSessions() error {
	sessions, err := awslib.BuildSessions()
	if err != nil {
		return errors.New("Unable to build AWS Sessions")
	}
	col.sessions = sessions
	return nil
}

// CheckCredentials tests the proper availability of AWS Credentials in the environment
func (col AWSCollector) CheckCredentials() bool {
	for _, sess := range col.sessions {
		_, err := sess.Config.Credentials.Get()
		if err != nil {
			return false
		}
		if sess.Config.Credentials.IsExpired() {
			return false
		}
	}
	return true
}

// CollectEC2 returns a concurrently collected EC2 inventory for all the regions
func (col AWSCollector) CollectEC2() (map[string][]*ec2.Instance, error) {
	instances := make(map[string][]*ec2.Instance)

	type instanceTuple struct {
		region    string
		instances []*ec2.Instance
	}

	instancesChan := make(chan instanceTuple, len(col.sessions))
	errChan := make(chan error, len(col.sessions))
	var wg sync.WaitGroup

	for region, sess := range col.sessions {
		wg.Add(1)
		go func(sess *session.Session, region string, instancesChan chan instanceTuple, errChan chan error) {
			defer wg.Done()
			chunk, err := CollectEC2PerSession(sess)

			if err != nil {
				errChan <- errors.New(fmt.Sprintf("Error while gathering %s: %v", region, err))
				return
			}

			// Ignore regions with no instances
			if chunk == nil {
				return
			}
			instancesChan <- instanceTuple{region, chunk}
		}(sess, region, instancesChan, errChan)
	}
	wg.Wait()
	close(instancesChan)
	close(errChan)

	if len(errChan) > 0 {
		return nil, errors.New(fmt.Sprintf("Failed to gather EC2 Data: %v", <-errChan))
	}

	for regionChunk := range instancesChan {
		instances[regionChunk.region] = regionChunk.instances
	}
	return instances, nil
}

// CollectEC2PerSession returns an EC2 inventory for a given session
func CollectEC2PerSession(sess *session.Session) ([]*ec2.Instance, error) {
	instances, err := awslib.GetAllInstances(sess)
	return instances, err
}
