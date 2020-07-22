package awslib

import (
        "github.com/aws/aws-sdk-go/aws/session"
        "github.com/aws/aws-sdk-go/service/cloudfront"
        "github.com/jpillora/backoff"
        "time"
)

// GetAllCDNInstances returns a complete list of instances for a given session
func GetAllCDNInstances(sess *session.Session) ([]*cloudfront.DistributionSummary, error) {

        b := &backoff.Backoff{
                //These are the defaults
                Min:    10 * time.Millisecond,
                Max:    1 * time.Second,
                Factor: 2,
                Jitter: false,
        }
        summary := make([]*cloudfront.DistributionSummary, 0)
        var nextPageExists = true
        request := &cloudfront.ListDistributionsInput{}

        cdn := cloudfront.New(sess)
        for nextPageExists {
                response, err := cdn.ListDistributions(request)

                if err != nil {
                        time.Sleep(b.Duration())
                } else {
                        result := response.DistributionList
                        for recordIndex := range response.DistributionList.Items {
                                summary = append(summary, response.DistributionList.Items[recordIndex])
                        }

                        if result.IsTruncated == nil || !*result.IsTruncated {
                                nextPageExists = false
                                break
                        }
                        // Setting next page.
                        request.Marker = result.NextMarker
                }
        }
        return summary, nil
}
