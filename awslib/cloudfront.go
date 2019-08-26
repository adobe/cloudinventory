package awslib

import (
	// "errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/jpillora/backoff"
)

func GetAllCloudFrontDistributions() ([]*cloudfront.DistributionList, error) {

	b := &backoff.Backoff{
		//These are the defaults
		Min:    10 * time.Millisecond,
		Max:    1 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	var nextPageExists = true
	var allDistributions []*cloudfront.DistributionList
	input := cloudfront.ListDistributionsInput{}
	cdn := cloudfront.New(session.New())

	fmt.Println("Now Starting......")
	for nextPageExists {
		result, err := cdn.ListDistributions(&input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				// Retry with backoff if RateExceeded
				if aerr.Code() == "RateExceeded" {
					time.Sleep(b.Duration())
					continue
				}
				return allDistributions, aerr
			}
			return allDistributions, err
		} else {
			allDistributions = append(allDistributions, result.DistributionList)
			fmt.Println(allDistributions)
			if result.DistributionList.IsTruncated == nil || !*result.DistributionList.IsTruncated {
				nextPageExists = false
				break
			}
			// Setting next page.
			input.Marker = result.DistributionList.Marker
		}
	}
	return allDistributions, nil
}
