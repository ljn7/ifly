package cmd

import (
	"fmt"

	"github.com/ljn7/ifly/cli/internal/config"
	"github.com/ljn7/ifly/cli/internal/paths"
	"github.com/ljn7/ifly/cli/internal/state"
	"github.com/spf13/cobra"
)

var modeProjectScope bool

var modeCmd = &cobra.Command{
	Use:   "mode <silent|minimal|normal|verbose|caveman>",
	Short: "Change the session verbosity mode",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mode := args[0]
		if err := config.ValidateMode(mode); err != nil {
			return err
		}
		statePath, err := modeStatePath()
		if err != nil {
			return err
		}
		st, err := state.Load(statePath)
		if err != nil {
			return err
		}
		st.Version = 1
		st.ActiveMode = mode
		if err := state.Save(statePath, st); err != nil {
			return err
		}
		scope := "global"
		if modeProjectScope {
			scope = "project"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "IFLy mode -> %s (%s)\n", mode, scope)
		return nil
	},
}

func modeStatePath() (string, error) {
	if modeProjectScope {
		return paths.ProjectStateFile()
	}
	return paths.StateFile()
}

func init() {
	modeCmd.Flags().BoolVar(&modeProjectScope, "project", false, "write .ifly-state.yaml in the current project")
	rootCmd.AddCommand(modeCmd)
}
