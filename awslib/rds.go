package awslib

import (
	"time"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/jpillora/backoff"
	"errors"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

//GetAllDBInstances resturns a complete list of DBInstances for a given session
func GetAllDBInstances(sess *session.Session) ([]*rds.DBInstance, error) {
	rdsc := rds.New(sess)
	allInstancesDone := false
	var allInstances []*rds.DBInstance
	input := rds.DescribeDBInstancesInput{}
	b := &backoff.Backoff{
		//These are the defaults
		Min:    10 * time.Millisecond,
		Max:    30 * time.Second,
		Factor: 2,
		Jitter: false,
	}
	for !allInstancesDone {
		// Describe instances with no filters
		result, err := rdsc.DescribeDBInstances(&input)
		if err != nil {
			if _, ok := err.(awserr.Error); ok {
				// Retry with backoff
				time.Sleep(b.Duration())
				continue
			}
			return allInstances, err
		}
		if err != nil {
			return nil, errors.New("Error Describing Instances")
		}
		allInstances = append(allInstances, result.DBInstances...)
		if result.Marker == nil {
			allInstancesDone = true
			break
		}
		input.SetMarker(*result.Marker)
	}
	return allInstances, nil
}
