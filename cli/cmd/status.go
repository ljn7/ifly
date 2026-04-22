package cmd

import (
	"fmt"
	"time"

	"github.com/ljn7/ifly/cli/internal/config"
	"github.com/ljn7/ifly/cli/internal/detect"
	"github.com/ljn7/ifly/cli/internal/egg"
	"github.com/ljn7/ifly/cli/internal/paths"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show active mode, guard level, and merged config summary",
	RunE: func(cmd *cobra.Command, args []string) error {
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
		env := config.LoadEnv()
		merged, violations := config.MergeRespectingLockdown(global, project, env)

		st, statePath, err := loadEffectiveState()
		if err != nil {
			return err
		}
		activeMode := merged.Mode
		if st.ActiveMode != "" {
			activeMode = st.ActiveMode
		}
		activeGuard := merged.Guard.Level
		if st.SessionGuardOverride != "" && !global.Guard.Lockdown {
			activeGuard = st.SessionGuardOverride
		}

		host := detect.Host()
		claudeDir, _ := detect.ClaudeConfigDir(host)

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "IFLy v%s\n", cliVersion)
		fmt.Fprintf(out, "mode: %s\n", nonEmpty(activeMode, "minimal"))
		fmt.Fprintf(out, "guard: %s\n", nonEmpty(activeGuard, "strict"))
		fmt.Fprintf(out, "lockdown: %v\n", global.Guard.Lockdown)
		fmt.Fprintf(out, "state: %s\n", statePath)
		fmt.Fprintf(out, "claude dir: %s\n", nonEmpty(claudeDir, "(not found)"))
		fmt.Fprintf(out, "project config: %s\n", nonEmpty(paths.ProjectConfigFile(), "(CLAUDE_PROJECT_DIR unset)"))
		fmt.Fprintln(out, "tools:", formatTools(merged.Guard.Tools))
		if n := len(merged.Guard.AdditionalDirs); n > 0 {
			fmt.Fprintf(out, "additional_dirs: %d entries\n", n)
		}
		if n := len(merged.Guard.AllowedNetwork); n > 0 {
			fmt.Fprintf(out, "allowed_network: %d entries\n", n)
		}
		for _, v := range violations {
			fmt.Fprintln(out, "lockdown:", v)
		}
		fmt.Fprintln(out, egg.Footer(cliVersion, merged.Telemetry.EasterEgg, int(time.Now().UnixNano()/int64(time.Millisecond))%1000))
		return nil
	},
}

func nonEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func formatTools(t config.Tools) string {
	labels := []struct {
		name string
		val  *bool
	}{
		{"bash", t.Bash}, {"edit", t.Edit}, {"write", t.Write},
		{"multi_edit", t.MultiEdit}, {"notebook_edit", t.NotebookEdit},
		{"read", t.Read}, {"glob", t.Glob}, {"grep", t.Grep},
		{"web_fetch", t.WebFetch}, {"web_search", t.WebSearch},
	}
	s := ""
	for i, l := range labels {
		state := "-"
		if l.val != nil {
			if *l.val {
				state = "on"
			} else {
				state = "off"
			}
		}
		if i > 0 {
			s += " "
		}
		s += l.name + "=" + state
	}
	return s
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
