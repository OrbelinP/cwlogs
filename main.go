package main

import (
	"runtime/debug"

	"github.com/OrbelinP/cwlogs/cmd/cwlogs"
	"github.com/alecthomas/kong"
)

func main() {
	cli, err := cwlogs.NewCLI()
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
