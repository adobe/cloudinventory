package awslib

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/jpillora/backoff"
	"time"
)

// GetAllInstances returns a complete list of instances for a given session
func GetAllHostedZones(sess *session.Session) ([]*route53.HostedZone, error) {

	b := &backoff.Backoff{
		//These are the defaults
		Min:    10 * time.Millisecond,
		Max:    1 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	zones := make([]*route53.HostedZone, 0);
	var nextPageExists = true
	request := &route53.ListHostedZonesInput{}

	r53 := route53.New(sess)

	for nextPageExists {
		response , err := r53.ListHostedZones(request)
		if err != nil {
			time.Sleep(b.Duration())
		}else {
			for recordIndex := range response.HostedZones {
				zones = append(zones ,response.HostedZones[recordIndex]);
			}
			if response.IsTruncated == nil || !*response.IsTruncated {
				nextPageExists = false
				break
			}
			// Setting next page.
			request.Marker = response.NextMarker;
		}
	}
	return zones, nil;
}