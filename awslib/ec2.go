package awslib

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// GetAllInstances returns a complete list of instances for a given session
func GetAllInstances(sess *session.Session) ([]*ec2.Instance, error) {
	ec2c := ec2.New(sess)
	allInstancesDone := false
	var allInstances []*ec2.Instance
	input := ec2.DescribeInstancesInput{}
	for !allInstancesDone {
		// Describe instances with no filters
		result, err := ec2c.DescribeInstances(&input)
		if err != nil {
			return allInstances, err
		}
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
