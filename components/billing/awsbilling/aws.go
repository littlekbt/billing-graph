package awsbilling

import (
	"fmt"

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

	now := time.Now()
	// yesterday
	startTime := time.Unix(now.Unix()-86400, int64(now.Nanosecond()))

	for _, metric := range billingMetricList.Metrics {
		v, err := getLatestValue(cw, metric.Dimensions, startTime, now)
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

func getLatestValue(cw *cloudwatch.CloudWatch, dimensions []*cloudwatch.Dimension, start time.Time, end time.Time) (float64, error) {
	statistics := []*string{aws.String("Maximum")}

	in := &cloudwatch.GetMetricStatisticsInput{
		Dimensions: dimensions,
		StartTime:  aws.Time(start),
		EndTime:    aws.Time(end),
		Namespace:  aws.String("AWS/Billing"),
		MetricName: aws.String("EstimatedCharges"),
		Period:     aws.Int64(3600),
		Statistics: statistics,
	}

	out, err := cw.GetMetricStatistics(in)

	if err != nil {
		return 0, err
	}

	datapoints := out.Datapoints
	if len(datapoints) == 0 {
		return 0, errors.New("fetched no datapoints")
	}

	var latest time.Time
	var latestIndex int

	fmt.Println(dimensions)
	fmt.Println(datapoints)
	for i, datapoint := range datapoints {
		if datapoint.Timestamp.After(latest) {
			latest = *datapoint.Timestamp
			latestIndex = i
		}
	}

	return *datapoints[latestIndex].Maximum, nil
}
