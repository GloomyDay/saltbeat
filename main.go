package main

import (
	"os"

	"github.com/elastic/beats/v7/libbeat/cmd"
	"github.com/elastic/beats/v7/libbeat/cmd/instance"

	"github.com/GloomyDay/saltbeat/beater"
)

var RootCmd = cmd.GenRootCmdWithSettings(beater.New, instance.Settings{Name: "saltbeat"})

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
