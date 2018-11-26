package awslib

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jpillora/backoff"
)

// GetAllInstances returns a complete list of instances for a given session
func GetAllInstances(sess *session.Session) ([]*ec2.Instance, error) {
	ec2c := ec2.New(sess)
	allInstancesDone := false
	var allInstances []*ec2.Instance
	input := ec2.DescribeInstancesInput{}
	b := &backoff.Backoff{
		//These are the defaults
		Min:    10 * time.Millisecond,
		Max:    30 * time.Second,
		Factor: 2,
		Jitter: false,
	}
	for !allInstancesDone {
		// Describe instances with no filters
		result, err := ec2c.DescribeInstances(&input)
		if err != nil {
			if _, ok := err.(awserr.Error); ok {
				// Retry with backoff
				time.Sleep(b.Duration())
				continue
			}
			return allInstances, err
		}
		b.Reset()
		for _, reservation := range result.Reservations {
			allInstances = append(allInstances, reservation.Instances...)
		}
		if result.NextToken == nil {
			allInstancesDone = true
			break
		}
		input.SetNextToken(*result.NextToken)
	}
	return allInstances, nil
}
