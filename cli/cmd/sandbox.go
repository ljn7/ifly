package cmd

import (
	"os"

	"github.com/ljn7/ifly/cli/internal/sandbox"
	"github.com/spf13/cobra"
)

var sandboxCmd = &cobra.Command{
	Use:                "sandbox [-- claude args...]",
	Short:              "Run claude inside a Linux filesystem namespace (Linux only)",
	Long:               "On Linux, wraps `claude` in a filesystem namespace via bwrap (or unshare fallback). System dirs are read-only bound; the current project is read-write bound. Provides real OS-level isolation that the plugin's advisory guard hook cannot match. Not supported on macOS/Windows.",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// DisableFlagParsing routes --help into args; handle it manually so
		// users still get cobra's usage screen.
		for _, a := range args {
			if a == "-h" || a == "--help" {
				return cmd.Help()
			}
		}
		argv := append([]string{"claude"}, args...)
		return sandbox.Run(cmd.OutOrStdout(), cmd.ErrOrStderr(), argv, os.Getenv("CLAUDE_PROJECT_DIR"))
	},
}

func init() {
	rootCmd.AddCommand(sandboxCmd)
}
