package cmd

import (
	"fmt"
	"os"

	"github.com/ljn7/ifly/cli/internal/config"
	"github.com/ljn7/ifly/cli/internal/paths"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configProjectScope bool

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Read and modify IFLy config",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print merged config (global + project + env)",
	RunE: func(cmd *cobra.Command, args []string) error {
		gp, err := paths.GlobalConfigFile()
		if err != nil {
			return err
		}
		g, err := config.LoadFile(gp)
		if err != nil {
			return err
		}
		p, err := config.LoadFile(paths.ProjectConfigFile())
		if err != nil {
			return err
		}
		merged, _ := config.MergeRespectingLockdown(g, p, config.LoadEnv())
		data, err := yaml.Marshal(merged)
		if err != nil {
			return err
		}
		cmd.Print(string(data))
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <dotted.key>",
	Short: "Read one key from merged config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gp, _ := paths.GlobalConfigFile()
		g, err := config.LoadFile(gp)
		if err != nil {
			return err
		}
		p, err := config.LoadFile(paths.ProjectConfigFile())
		if err != nil {
			return err
		}
		merged, _ := config.MergeRespectingLockdown(g, p, config.LoadEnv())
		v, err := getDotted(merged, args[0])
		if err != nil {
			return err
		}
		cmd.Println(v)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <dotted.key> <value>",
	Short: "Write one key to global config (or project with --project)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, val := args[0], args[1]
		if err := validateSet(key, val); err != nil {
			return err
		}
		target, err := paths.GlobalConfigFile()
		if err != nil {
			return err
		}
		if configProjectScope {
			if p := paths.ProjectConfigFile(); p != "" {
				target = p
			} else {
				return fmt.Errorf("--project requires CLAUDE_PROJECT_DIR")
			}
		}
		c, err := config.LoadFile(target)
		if err != nil {
			return err
		}
		if err := setDotted(&c, key, val); err != nil {
			return err
		}
		if c.Version == 0 {
			c.Version = 1
		}
		data, err := yaml.Marshal(c)
		if err != nil {
			return err
		}
		if err := paths.EnsureDir(filepathDir(target)); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
		cmd.Printf("set %s = %s in %s\n", key, val, target)
		return nil
	},
}

func init() {
	configSetCmd.Flags().BoolVar(&configProjectScope, "project", false, "write to .ifly.yaml instead of global")
	configCmd.AddCommand(configShowCmd, configGetCmd, configSetCmd)
	rootCmd.AddCommand(configCmd)
}
