package cmd

import (
	"github.com/spf13/cobra"
)

var (
	cliVersion = "dev"
	repoOwner  = "ljn7"
	repoName   = "ifly"
)

var rootCmd = &cobra.Command{
	Use:           "ifly",
	Short:         "IFLy — Claude Code power-pack",
	Long:          "IFLy installs and manages the ifly Claude Code plugin, exposes merged config, and on Linux offers a namespace-based sandbox.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// SetVersion is called from main so ldflags-injected version reaches commands.
func SetVersion(v string) { cliVersion = v }

// SetRepo lets tests override the release repo.
func SetRepo(owner, name string) { repoOwner, repoName = owner, name }

func Execute() error { return rootCmd.Execute() }
