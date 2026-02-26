package cwlogs

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/OrbelinP/cwlogs/cmd/cwlogs/mocks"
	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCLI_Run(t *testing.T) {
	t.Run("using --last with no history", func(t *testing.T) {
		// given
		outBuf := new(bytes.Buffer)
		tmp := t.TempDir()
		cli, err := NewCLI(nil, outBuf, tmp)
		require.NoError(t, err)
		parser, err := kong.New(cli)
		require.NoError(t, err)

		history, err := LoadHistory(tmp)
		require.NoError(t, err)
		require.Len(t, history.LogGroups, 0)

		// when
		_, err = parser.Parse([]string{"--last"})
		require.NoError(t, err)

		require.NoError(t, cli.Run())

		// then
		assert.Equal(t, "No log groups in history", outBuf.String())
	})

	t.Run("using --last with history", func(t *testing.T) {
		// given
		ctrl := gomock.NewController(t)
		cw := mocks.NewMockCloudWatchClient(ctrl)
		out := &cloudwatchlogs.FilterLogEventsOutput{
			NextToken: aws.String("next-token"),
		}
		cw.EXPECT().
			FilterLogEvents(gomock.Any(), eqFilterLogEventsInputMatcher{&cloudwatchlogs.FilterLogEventsInput{
				LogGroupName: aws.String("detail three"),
			}}, gomock.Any()).
			Return(out, nil)

		cw.EXPECT().
			FilterLogEvents(gomock.Any(), eqFilterLogEventsInputMatcher{&cloudwatchlogs.FilterLogEventsInput{
				LogGroupName: aws.String("detail three"),
				NextToken:    aws.String("next-token"),
			}}, gomock.Any()).
			Return(&cloudwatchlogs.FilterLogEventsOutput{}, nil)

		outBuf := new(bytes.Buffer)
		tmp := t.TempDir()
		cli, err := NewCLI(cw, outBuf, tmp)
		require.NoError(t, err)
		parser, err := kong.New(cli)
		require.NoError(t, err)

		detailOne := LogGroupDetails{FullName: "detail one", ShortName: "one"}
		detailTwo := LogGroupDetails{FullName: "detail two", ShortName: "two"}
		detailThree := LogGroupDetails{FullName: "detail three", ShortName: "three"}

		require.NoError(t, AddToHistory(detailOne, tmp))
		require.NoError(t, AddToHistory(detailTwo, tmp))
		require.NoError(t, AddToHistory(detailThree, tmp))

		history, err := LoadHistory(tmp)
		require.NoError(t, err)
		require.Len(t, history.LogGroups, 3)
		require.Equal(t, detailThree, history.LogGroups[0])
		require.Equal(t, detailTwo, history.LogGroups[1])
		require.Equal(t, detailOne, history.LogGroups[2])

		// when
		_, err = parser.Parse([]string{"--last", "--timeout", "0s"})
		require.NoError(t, err)

		require.NoError(t, cli.Run())

		// then
		assert.Contains(t, outBuf.String(), "Selected: detail three")
	})
}

type eqFilterLogEventsInputMatcher struct {
	want *cloudwatchlogs.FilterLogEventsInput
}

func (e eqFilterLogEventsInputMatcher) Matches(x any) bool {
	arg, ok := x.(*cloudwatchlogs.FilterLogEventsInput)
	if !ok {
		return false
	}

	return reflect.DeepEqual(e.want.LogGroupName, arg.LogGroupName) &&
		reflect.DeepEqual(e.want.NextToken, arg.NextToken)
}

func (e eqFilterLogEventsInputMatcher) String() string {
	return fmt.Sprintln(e.want)
}
