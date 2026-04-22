package main

import (
	"fmt"
	"os"

	"github.com/ljn7/ifly/cli/cmd"
)

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.SetPluginFS(PluginFS)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "ifly:", err)
		os.Exit(1)
	}
}
