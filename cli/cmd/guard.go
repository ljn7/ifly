package cmd

import (
	"fmt"

	"github.com/ljn7/ifly/cli/internal/config"
	"github.com/ljn7/ifly/cli/internal/paths"
	"github.com/ljn7/ifly/cli/internal/state"
	"github.com/spf13/cobra"
)

var guardProjectScope bool

var guardCmd = &cobra.Command{
	Use:   "guard [strict|project|open|off|status]",
	Short: "Show or change the session guard override",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || args[0] == "status" {
			return printGuardStatus(cmd)
		}
		return setGuardOverride(cmd, args[0])
	},
}

func printGuardStatus(cmd *cobra.Command) error {
	globalPath, err := paths.GlobalConfigFile()
	if err != nil {
		return err
	}
	global, err := config.LoadFile(globalPath)
	if err != nil {
		return err
	}
	project, err := config.LoadFile(paths.ProjectConfigFile())
	if err != nil {
		return err
	}
	merged, violations := config.MergeRespectingLockdown(global, project, config.LoadEnv())

	st, statePath, err := loadEffectiveState()
	if err != nil {
		return err
	}

	level := merged.Guard.Level
	if st.SessionGuardOverride != "" {
		if !global.Guard.Lockdown || config.LevelRank(st.SessionGuardOverride) >= config.LevelRank(level) {
			level = st.SessionGuardOverride
		}
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "guard level: %s\n", nonEmpty(level, "strict"))
	fmt.Fprintf(out, "session override: %s\n", nonEmpty(st.SessionGuardOverride, "(none)"))
	fmt.Fprintf(out, "state: %s\n", statePath)
	fmt.Fprintf(out, "lockdown: %v\n", global.Guard.Lockdown)
	if len(merged.Guard.AdditionalDirs) == 0 {
		fmt.Fprintln(out, "additional_dirs: []")
	} else {
		fmt.Fprintln(out, "additional_dirs:")
		for _, dir := range merged.Guard.AdditionalDirs {
			fmt.Fprintf(out, "  - %s\n", dir)
		}
	}
	for _, v := range violations {
		fmt.Fprintln(out, "lockdown:", v)
	}
	return nil
}

func setGuardOverride(cmd *cobra.Command, level string) error {
	if err := config.ValidateLevel(level); err != nil {
		return err
	}

	globalPath, err := paths.GlobalConfigFile()
	if err != nil {
		return err
	}
	global, err := config.LoadFile(globalPath)
	if err != nil {
		return err
	}
	baseline := nonEmpty(global.Guard.Level, "strict")
	if global.Guard.Lockdown && config.LevelRank(level) < config.LevelRank(baseline) {
		return fmt.Errorf("IFLy: lockdown enabled - cannot set session override looser than global guard level")
	}

	statePath, err := guardStatePath()
	if err != nil {
		return err
	}
	st, err := state.Load(statePath)
	if err != nil {
		return err
	}
	st.Version = 1
	st.SessionGuardOverride = level
	if err := state.Save(statePath, st); err != nil {
		return err
	}
	scope := "global"
	if guardProjectScope {
		scope = "project"
	}
	fmt.Fprintf(cmd.OutOrStdout(), "IFLy guard -> %s (%s)\n", level, scope)
	return nil
}

func guardStatePath() (string, error) {
	if guardProjectScope {
		return paths.ProjectStateFile()
	}
	return paths.StateFile()
}

func loadEffectiveState() (state.State, string, error) {
	globalPath, err := paths.StateFile()
	if err != nil {
		return state.State{}, "", err
	}
	global, err := state.Load(globalPath)
	if err != nil {
		return state.State{}, "", err
	}
	projectPath, err := paths.ProjectStateFile()
	if err != nil {
		return state.State{}, "", err
	}
	project, err := state.Load(projectPath)
	if err != nil {
		return state.State{}, "", err
	}
	if project.ActiveMode != "" || project.SessionGuardOverride != "" {
		return state.Merge(global, project), projectPath, nil
	}
	return global, globalPath, nil
}

func init() {
	guardCmd.Flags().BoolVar(&guardProjectScope, "project", false, "write .ifly-state.yaml in the current project")
	rootCmd.AddCommand(guardCmd)
}
