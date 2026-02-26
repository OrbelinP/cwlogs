package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"charm.land/lipgloss/v2"
	"github.com/OrbelinP/cwlogs/cmd/cwlogs"
	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("loading default config: %v", err))
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(fmt.Sprintf("loading user config dir: %v", err))
	}

	cli, err := cwlogs.NewCLI(cloudwatchlogs.NewFromConfig(cfg), os.Stdout, configDir)
	if err != nil {
		panic(err)
	}

	kongCtx := kong.Parse(cli,
		kong.Name("cwlogs"),
		kong.Description("A TUI tool to list and tail CloudWatch log groups"),
		kong.Configuration(kongyaml.Loader, "~/.cwlogs.yaml"),
		kong.UsageOnError(),
		kong.Vars{"version": getVersion()},
	)
	err = kongCtx.Run()
	if err != nil {
		msg := lipgloss.NewStyle().Foreground(lipgloss.Red).Render(err.Error())
		fmt.Fprintln(os.Stderr, msg)
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
