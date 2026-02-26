package main

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

type CLI struct {
	cw *cloudwatchlogs.Client

	Pattern *string       `name:"pattern" short:"p" help:"CloudWatch describe log groups pattern"`
	Since   time.Duration `name:"since" short:"s" default:"0h" help:"CloudWatch describe log groups since"`

	Timeout time.Duration `name:"timeout" short:"t" default:"1h" help:"Timeout for tailing selected log group"`

	Last bool `name:"last" default:"false" help:"Select most recently selected log group"`

	Version kong.VersionFlag `name:"version" help:"Show version information"`
}

type LogGroupDetails struct {
	FullName  string `json:"fullName"`
	ShortName string `json:"shortName"`
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	kongCtx := kong.Parse(&CLI{
		cw: cloudwatchlogs.NewFromConfig(cfg),
	},
		kong.Name("cwlogs"),
		kong.Description("A TUI tool to list and tail CloudWatch log groups"),
		kong.UsageOnError(),
		kong.Vars{"version": getVersion()},
	)
	err = kongCtx.Run()
	kongCtx.FatalIfErrorf(err)
}

func (cli *CLI) Run() error {
	selected, err := cli.selectALogGroup()
	if err != nil {
		return fmt.Errorf("selecting a log group: %w", err)
	}

	if selected == nil {
		return nil
	}

	detail := *selected
	err = AddToHistory(detail)
	if err != nil {
		return fmt.Errorf("adding log group to history: %w", err)
	}

	finalMsg := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("#438f39")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#438f39")).
		Render(fmt.Sprintf("Selected: %s", detail.FullName))

	fmt.Println(finalMsg)

	err = cli.tailLogs(detail.FullName)
	if err != nil {
		return fmt.Errorf("tailing logs: %w", err)
	}

	return nil
}

func (cli *CLI) selectALogGroup() (*LogGroupDetails, error) {
	if cli.Last {
		h, err := LoadHistory()
		if err != nil {
			return nil, fmt.Errorf("loading history: %w", err)
		}

		return &h.LogGroups[0], nil
	}

	logGroups, err := cli.listLogGroups()
	if err != nil {
		return nil, fmt.Errorf("listing log groups: %w", err)
	}

	m := newSelectModel(logGroups)
	returnModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, fmt.Errorf("running tea program to view logs: %w", err)
	}

	m, ok := returnModel.(selectModel)
	if !ok {
		return nil, nil
	}

	if m.selected == nil {
		return nil, nil
	}

	detail, ok := (*m.selected).(LogGroupDetails)
	if !ok {
		return nil, nil
	}

	return &detail, nil
}

func (cli *CLI) listLogGroups() ([]LogGroupDetails, error) {
	in := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePattern: cli.Pattern,
	}

	first := true
	result := make([]LogGroupDetails, 0)
	for first || in.NextToken != nil {
		first = false

		out, err := cli.cw.DescribeLogGroups(context.Background(), in)
		if err != nil {
			return nil, fmt.Errorf("describing log groups: %w", err)
		}

		for _, g := range out.LogGroups {
			result = append(result, toLogGroupDetails(*g.LogGroupName))
		}

		in.NextToken = out.NextToken
	}

	return result, nil
}

func (cli *CLI) tailLogs(lgName string) error {
	startTime := time.Now().Add(-1 * cli.Since).UnixMilli()
	in := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(lgName),
		StartTime:    aws.Int64(startTime),
	}

	ctx, cancel := context.WithTimeout(context.Background(), cli.Timeout)
	defer cancel()

	seen := make(map[string]int64)

	for {
		paginator := cloudwatchlogs.NewFilterLogEventsPaginator(cli.cw, in)

		for paginator.HasMorePages() {
			out, err := paginator.NextPage(ctx)
			if err != nil {
				return fmt.Errorf("getting next page of log events: %w", err)
			}

			for _, e := range out.Events {
				id := *e.EventId
				if _, ok := seen[id]; ok {
					continue
				}

				seen[id] = *e.Timestamp
				t := time.UnixMilli(*e.Timestamp)
				fmt.Printf("[%s] %s\n", t.Format(time.RFC3339), *e.Message)

				in.StartTime = aws.Int64(max(*in.StartTime, *e.Timestamp))
			}
		}

		// remove any messages older than 5s from the seen map
		cutoff := *in.StartTime - 5_000
		for id, t := range seen {
			if t < cutoff {
				delete(seen, id)
			}
		}

		select {
		case <-ctx.Done():
			fmt.Printf("Provided timeout duration (%s) elapsed, stopping...", cli.Timeout)
			return nil
		case <-time.After(1 * time.Second):
		}
	}
}

func (lg LogGroupDetails) FilterValue() string {
	return lg.FullName
}

func (lg LogGroupDetails) Title() string {
	return lg.ShortName
}

func (lg LogGroupDetails) Description() string {
	return lg.ShortName
}

func toLogGroupDetails(name string) LogGroupDetails {
	shortName := name
	if len(shortName) > 100 {
		shortName = fmt.Sprintf("%s...", shortName[:100])
	}

	return LogGroupDetails{
		FullName:  name,
		ShortName: shortName,
	}
}

func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(unknown)"
	}

	if info.Main.Version != "" {
		return info.Main.Version
	}

	return "(unknown)"
}
