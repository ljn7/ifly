package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ljn7/ifly/cli/internal/config"
	"github.com/ljn7/ifly/cli/internal/paths"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type blockPreset struct {
	Name        string
	Category    string
	Modes       string
	Description string
	Patterns    []string
}

var blockPresets = []blockPreset{
	{
		Name:        "git-danger",
		Category:    "git",
		Modes:       "all",
		Description: "block destructive git history/worktree operations",
		Patterns: []string{
			"git reset --hard",
			"git clean -fd",
			"git clean -fdx",
			"git push --force",
			"git push --force-with-lease",
			"git checkout -- .",
			"git restore .",
		},
	},
	{
		Name:        "archive-overwrite",
		Category:    "archive",
		Modes:       "all",
		Description: "block archive extraction forms that commonly overwrite or escape directories",
		Patterns: []string{
			"tar --overwrite",
			"tar -C /",
			"tar -xPf",
			"tar --absolute-names",
			"bsdtar --overwrite",
			"unzip -o",
		},
	},
	{
		Name:        "shell-wrapper",
		Category:    "shell",
		Modes:       "all",
		Description: "block nested shell wrappers often used to hide final commands",
		Patterns: []string{
			"bash -c",
			"sh -c",
			"powershell -Command",
			"pwsh -Command",
			"cmd /c",
		},
	},
	{
		Name:        "filesystem-danger",
		Category:    "filesystem",
		Modes:       "all",
		Description: "block broad recursive permission and deletion patterns everywhere",
		Patterns: []string{
			"rm -rf",
			"rmdir /s",
			"Remove-Item -Recurse",
			"chmod -R",
			"chown -R",
		},
	},
}

var blockProjectScope bool

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Manage guard blocked command patterns",
	Long:  "Manage guard.blocked_commands. These literal substrings are blocked at every guard level, including off.",
}

var blockListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"get"},
	Short:   "Show active blocked command patterns",
	RunE: func(cmd *cobra.Command, args []string) error {
		gp, err := paths.GlobalConfigFile()
		if err != nil {
			return err
		}
		global, err := config.LoadFile(gp)
		if err != nil {
			return err
		}
		project, err := config.LoadFile(paths.ProjectConfigFile())
		if err != nil {
			return err
		}
		merged, violations := config.MergeRespectingLockdown(global, project, config.LoadEnv())
		if len(violations) > 0 {
			for _, v := range violations {
				cmd.Printf("warning: %s\n", v)
			}
		}
		if len(merged.Guard.BlockedCommands) == 0 {
			cmd.Println("blocked commands: []")
			return nil
		}
		cmd.Println("blocked commands:")
		for _, p := range merged.Guard.BlockedCommands {
			cmd.Printf("  - %s%s\n", p, presetTag(p))
		}
		cmd.Println("scope: literal substring match; active in all guard levels")
		return nil
	},
}

var blockPresetsCmd = &cobra.Command{
	Use:   "presets",
	Short: "List built-in blocked command presets",
	Run: func(cmd *cobra.Command, args []string) {
		for _, p := range blockPresets {
			cmd.Printf("%s [%s, %s] - %s\n", p.Name, p.Category, p.Modes, p.Description)
			for _, pattern := range p.Patterns {
				cmd.Printf("  - %s\n", pattern)
			}
		}
	},
}

var blockAddCmd = &cobra.Command{
	Use:   "add <literal-substring>",
	Short: "Add a custom blocked command pattern",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, c, err := loadWritableBlockConfig()
		if err != nil {
			return err
		}
		c.Guard.BlockedCommands = appendUnique(c.Guard.BlockedCommands, args[0])
		if err := saveBlockConfig(target, c); err != nil {
			return err
		}
		cmd.Printf("added blocked command %q to %s\n", args[0], target)
		return nil
	},
}

var blockPresetCmd = &cobra.Command{
	Use:   "preset <name>",
	Short: "Add a built-in preset to blocked_commands",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, ok := findBlockPreset(args[0])
		if !ok {
			return fmt.Errorf("unknown preset %q; run `ifly block presets`", args[0])
		}
		target, c, err := loadWritableBlockConfig()
		if err != nil {
			return err
		}
		before := len(c.Guard.BlockedCommands)
		c.Guard.BlockedCommands = appendUnique(c.Guard.BlockedCommands, p.Patterns...)
		if err := saveBlockConfig(target, c); err != nil {
			return err
		}
		cmd.Printf("added preset %s (%d new patterns) to %s\n", p.Name, len(c.Guard.BlockedCommands)-before, target)
		return nil
	},
}

var blockCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List preset categories",
	Run: func(cmd *cobra.Command, args []string) {
		seen := map[string]struct{}{}
		for _, p := range blockPresets {
			seen[p.Category] = struct{}{}
		}
		cats := make([]string, 0, len(seen))
		for c := range seen {
			cats = append(cats, c)
		}
		sort.Strings(cats)
		for _, c := range cats {
			cmd.Println(c)
		}
	},
}

func loadWritableBlockConfig() (string, config.Config, error) {
	target, err := paths.GlobalConfigFile()
	if err != nil {
		return "", config.Config{}, err
	}
	if blockProjectScope {
		if p := paths.ProjectConfigFile(); p != "" {
			target = p
		} else {
			return "", config.Config{}, fmt.Errorf("--project requires CLAUDE_PROJECT_DIR")
		}
	}
	c, err := config.LoadFile(target)
	return target, c, err
}

func saveBlockConfig(target string, c config.Config) error {
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
	return os.WriteFile(target, data, 0o644)
}

func appendUnique(values []string, add ...string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values)+len(add))
	for _, v := range append(values, add...) {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func findBlockPreset(name string) (blockPreset, bool) {
	for _, p := range blockPresets {
		if p.Name == name {
			return p, true
		}
	}
	return blockPreset{}, false
}

func presetTag(pattern string) string {
	for _, p := range blockPresets {
		for _, candidate := range p.Patterns {
			if candidate == pattern {
				return " (" + p.Name + ")"
			}
		}
	}
	if strings.TrimSpace(pattern) == "" {
		return " (empty ignored by hook)"
	}
	return ""
}

func init() {
	blockCmd.PersistentFlags().BoolVar(&blockProjectScope, "project", false, "write/read project config where supported")
	blockCmd.AddCommand(blockListCmd, blockPresetsCmd, blockAddCmd, blockPresetCmd, blockCategoriesCmd)
	rootCmd.AddCommand(blockCmd)
}
