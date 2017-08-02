package awsbilling

import (
	"errors"
	"fmt"
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

	credentials *credentials.Credentials
	cloudWatch  *cloudwatch.CloudWatch
}

// Get gets latest billing values.
func (ab AWS) Get() int {
	if ab.AccessKeyID != "" && ab.SecretAccessKey != "" {
		ab.credentials = credentials.NewStaticCredentials(ab.AccessKeyID, ab.SecretAccessKey, "")
	}

	ab.cloudWatch = cloudwatch.New(session.New(
		&aws.Config{
			Credentials: ab.credentials,
			Region:      aws.String(ab.Region),
		},
	))

	billingMetricList, _ := ab.cloudWatch.ListMetrics(&cloudwatch.ListMetricsInput{Namespace: aws.String("AWS/Billing")})

	targets := make([]string, 0)
	for _, metric := range billingMetricList.Metrics {
		for _, dimension := range metric.Dimensions {
			if *dimension.Name == "ServiceName" {
				targets = append(targets, *dimension.Value)
			}
		}
	}

	baseDimension := []*cloudwatch.Dimension{&cloudwatch.Dimension{
		Name:  aws.String("Currency"),
		Value: aws.String(ab.Currency),
	}}

	for _, metricName := range targets {
		var dimentions []*cloudwatch.Dimension
		if metricName == "All" {
			dimentions = baseDimension
		} else {
			dimentions = append(
				[]*cloudwatch.Dimension{baseDimension[0]},
				[]*cloudwatch.Dimension{&cloudwatch.Dimension{
					Name:  aws.String("ServiceName"),
					Value: aws.String(metricName),
				}}...)
		}

		fmt.Println(metricName)
		fmt.Println(getLatestValue(ab.cloudWatch, dimentions))
	}

	return 0
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
