package awslib

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
)

// GetAllRegions returns all regions for AWS except US-Gov
func GetAllRegions(china bool) []string {
	awsRegions := endpoints.AwsPartition().Regions()
	var regions []string
	for _, r := range awsRegions {
		regions = append(regions, r.ID())
	}
	if china {
		cnRegions := endpoints.AwsCnPartition().Regions()
		for _, r := range cnRegions {
			regions = append(regions, r.ID())
		}
	}
	return regions
}

// BuildSessions returns a map of sessions for each region
func BuildSessions(china bool) (map[string]*session.Session, error) {
	sessions := make(map[string]*session.Session)
	var errMain error
	regions := GetAllRegions(china)
	for _, region := range regions {
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(region),
		})
		if err != nil {
			errMain = err
		}
		sessions[region] = sess
	}
	return sessions, errMain
}
