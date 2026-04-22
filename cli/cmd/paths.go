package cmd

import (
	"os"
	"path/filepath"

	"github.com/ljn7/ifly/cli/internal/detect"
	"github.com/ljn7/ifly/cli/internal/paths"
	"github.com/spf13/cobra"
)

var pathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "Show Claude and IFLy config/plugin paths",
	RunE: func(cmd *cobra.Command, args []string) error {
		host := detect.Host()
		candidates, err := detect.ClaudeConfigCandidates(host)
		if err != nil {
			return err
		}
		preferred := candidates[0]
		globalConfig, err := paths.GlobalConfigFile()
		if err != nil {
			return err
		}
		stateFile, err := paths.StateFile()
		if err != nil {
			return err
		}

		cmd.Printf("host: %s/%s\n", host.OS, host.Arch)
		cmd.Printf("ifly config: %s\n", globalConfig)
		cmd.Printf("ifly state: %s\n", stateFile)
		cmd.Printf("project config: %s\n", nonEmpty(paths.ProjectConfigFile(), "(CLAUDE_PROJECT_DIR unset)"))
		cmd.Println("claude config candidates:")
		for i, c := range candidates {
			marker := " "
			if i == 0 {
				marker = "*"
			}
			cmd.Printf("  %s %s [%s]\n", marker, c, pathState(c))
		}
		pluginDir := filepath.Join(preferred, "plugins", "ifly")
		cmd.Printf("ifly plugin target: %s [%s]\n", pluginDir, pathState(pluginDir))
		return nil
	},
}

func pathState(p string) string {
	info, err := os.Stat(p)
	if err != nil {
		return "missing"
	}
	if info.IsDir() {
		return "dir"
	}
	return "file"
}

func init() {
	rootCmd.AddCommand(pathsCmd)
}
