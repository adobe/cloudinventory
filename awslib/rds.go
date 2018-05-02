package awslib

import (
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
	for !allInstancesDone {
		// Describe instances with no filters
		result, err := rdsc.DescribeDBInstances(&input)
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
