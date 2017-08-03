package awsbilling

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// AWS inplemnted Billing.
type AWS struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Currency        string
}

// Get gets latest billing values.
func (ab AWS) Get() (map[string]float64, error) {

	if ab.AccessKeyID == "" || ab.SecretAccessKey == "" {
		return nil, errors.New("error: AccessKeyID or SecretAccessKey isn't defined")
	}

	credential := credentials.NewStaticCredentials(ab.AccessKeyID, ab.SecretAccessKey, "")

	cw := cloudwatch.New(session.New(
		&aws.Config{
			Credentials: credential,
			Region:      aws.String(ab.Region),
		},
	))

	// get target metric names
	billingMetricList, _ := cw.ListMetrics(&cloudwatch.ListMetricsInput{Namespace: aws.String("AWS/Billing")})
	r := make(map[string]float64)
	for _, metric := range billingMetricList.Metrics {
		v, err := getLatestValue(cw, metric.Dimensions)
		if err != nil {
			continue
		}
		for _, d := range metric.Dimensions {
			if *d.Name == "ServiceName" {
				r[*d.Value] = v
			}
		}
	}

	return r, nil
}

func getLatestValue(cloudWatch *cloudwatch.CloudWatch, dimensions []*cloudwatch.Dimension) (float64, error) {
	now := time.Now()

	startTime := time.Unix(now.Unix()-86400, int64(now.Nanosecond()))

	statistics := []*string{aws.String("Maximum")}

	in := &cloudwatch.GetMetricStatisticsInput{
		Dimensions: dimensions,
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(now),
		Namespace:  aws.String("AWS/Billing"),
		MetricName: aws.String("EstimatedCharges"),
		Period:     aws.Int64(3600),
		Statistics: statistics,
	}

	out, err := cloudWatch.GetMetricStatistics(in)

	if err != nil {
		return 0, nil
	}

	datapoints := out.Datapoints
	if len(datapoints) == 0 {
		return 0, errors.New("fetched no datapoints")
	}

	var latest time.Time
	var latestIndex int

	for i, datapoint := range datapoints {
		if datapoint.Timestamp.After(latest) {
			latest = *datapoint.Timestamp
			latestIndex = i
		}
	}

	return *datapoints[latestIndex].Maximum, nil
}
