package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ljn7/ifly/cli/internal/paths"
	"github.com/spf13/cobra"
)

var initForce bool

const iflyYAMLTemplate = `# IFLy project config. See docs/cli.md for full reference.
version: 1

# mode: silent | minimal | normal | verbose | caveman
# (omit to inherit global setting)
# mode: minimal

guard:
  # level: strict | project | open | off
  # level: project

  additional_dirs: []
  blocked_commands: []
  allowed_network: []
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Write a starter .ifly.yaml into the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		project := paths.ProjectConfigFile()
		if project == "" {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			project = filepath.Join(wd, ".ifly.yaml")
		}
		if _, err := os.Stat(project); err == nil && !initForce {
			return fmt.Errorf("%s already exists; pass --force to overwrite", project)
		} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		if err := os.WriteFile(project, []byte(iflyYAMLTemplate), 0o644); err != nil {
			return err
		}
		cmd.Printf("wrote %s\n", project)
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing .ifly.yaml")
	rootCmd.AddCommand(initCmd)
}
