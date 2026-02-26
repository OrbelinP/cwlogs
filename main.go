package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/OrbelinP/cwlogs/cmd/cwlogs"
	"github.com/alecthomas/kong"
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
		kong.UsageOnError(),
		kong.Vars{"version": getVersion()},
	)
	err = kongCtx.Run()
	kongCtx.FatalIfErrorf(err)
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
