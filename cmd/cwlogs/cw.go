package cwlogs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

//go:generate mockgen -source $GOFILE -destination=mocks/${GOFILE} -package mocks -typed

type CloudWatchClient interface {
	FilterLogEvents(
		ctx context.Context,
		in *cloudwatchlogs.FilterLogEventsInput,
		opts ...func(*cloudwatchlogs.Options),
	) (*cloudwatchlogs.FilterLogEventsOutput, error)

	DescribeLogGroups(
		ctx context.Context,
		in *cloudwatchlogs.DescribeLogGroupsInput,
		opts ...func(options *cloudwatchlogs.Options),
	) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
}
