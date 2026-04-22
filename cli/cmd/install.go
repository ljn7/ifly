package cmd

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ljn7/ifly/cli/internal/config"
	"github.com/ljn7/ifly/cli/internal/detect"
	"github.com/ljn7/ifly/cli/internal/installer"
	"github.com/ljn7/ifly/cli/internal/paths"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// pluginFS holds the embedded plugin tree. main.go sets this via SetPluginFS
// so the cmd package doesn't need to import the main package's embed var.
var pluginFS fs.FS

func SetPluginFS(f fs.FS) { pluginFS = f }

type InstallOpts struct {
	PluginFS  fs.FS
	ClaudeDir string
	Mode      string
	Guard     string
	Lockdown  bool
	Overwrite bool
}

var (
	installOverwrite      bool
	installNonInteractive bool
	installModeFlag       string
	installGuardFlag      string
	installLockdownFlag   bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the ifly plugin into Claude Code",
	RunE: func(cmd *cobra.Command, args []string) error {
		host := detect.Host()
		claudeDir, err := detect.ClaudeConfigDirPreferred(host)
		if err != nil {
			return err
		}

		if installNonInteractive {
			opts := InstallOpts{
				PluginFS:  pluginFS,
				ClaudeDir: claudeDir,
				Mode:      installModeFlag,
				Guard:     installGuardFlag,
				Lockdown:  installLockdownFlag,
				Overwrite: installOverwrite,
			}
			if !opts.Overwrite {
				ok, err := confirmOverwriteIfPopulated(cmd.InOrStdin(), cmd.OutOrStdout(), filepath.Join(claudeDir, "plugins", "ifly"))
				if err != nil {
					return err
				}
				opts.Overwrite = ok
			}
			return runInstall(cmd.OutOrStdout(), opts)
		}

		m := newInstallModel(claudeDir)
		p := tea.NewProgram(m)
		final, err := p.Run()
		if err != nil {
			return err
		}
		fm := final.(installModel)
		if fm.aborted {
			return fmt.Errorf("install aborted")
		}
		opts := fm.opts(pluginFS, installOverwrite)
		if !opts.Overwrite {
			ok, err := confirmOverwriteIfPopulated(cmd.InOrStdin(), cmd.OutOrStdout(), filepath.Join(claudeDir, "plugins", "ifly"))
			if err != nil {
				return err
			}
			opts.Overwrite = ok
		}
		return runInstall(cmd.OutOrStdout(), opts)
	},
}

func confirmOverwriteIfPopulated(in io.Reader, out io.Writer, pluginDest string) (bool, error) {
	entries, err := os.ReadDir(pluginDest)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if len(entries) == 0 {
		return false, nil
	}
	fmt.Fprintf(out, "%s already contains plugin files. Overwrite? [y/N]: ", pluginDest)
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && len(line) == 0 {
		return false, fmt.Errorf("%s already populated; pass --overwrite to replace", pluginDest)
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	if answer == "y" || answer == "yes" {
		return true, nil
	}
	return false, fmt.Errorf("install aborted; %s already populated", pluginDest)
}

func runInstall(out io.Writer, o InstallOpts) error {
	if err := config.ValidateMode(o.Mode); err != nil {
		return err
	}
	if err := config.ValidateLevel(o.Guard); err != nil {
		return err
	}
	if !detect.ClaudeBinaryPresent() {
		fmt.Fprintln(out, "warning: `claude` not found on PATH; plugin will be unused until it is installed")
	}
	if detect.Host().OS == "windows" && !detect.BashLikelyAvailable() {
		fmt.Fprintln(out, "warning: Windows install without Git Bash/WSL. Hooks are bash scripts and will NOT run. Install Git for Windows or enable WSL before relying on guards.")
	}
	pluginDest := filepath.Join(o.ClaudeDir, "plugins", "ifly")
	if err := installer.Install(o.PluginFS, installer.Options{Dest: pluginDest, Overwrite: o.Overwrite}); err != nil {
		return err
	}
	report, _ := installer.Verify(pluginDest)
	if !report.OK {
		return fmt.Errorf("post-install verify failed: %v", report.Failures)
	}

	globalPath, err := paths.GlobalConfigFile()
	if err != nil {
		return err
	}
	if err := paths.EnsureDir(filepath.Dir(globalPath)); err != nil {
		return err
	}
	cfg := config.Config{
		Version: 1,
		Mode:    o.Mode,
		Guard:   config.Guard{Level: o.Guard, Lockdown: o.Lockdown},
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(globalPath, data, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(out, "installed plugin to %s\n", pluginDest)
	fmt.Fprintf(out, "wrote global config %s\n", globalPath)
	fmt.Fprintln(out, "next: launch claude and run /ifly:status")
	return nil
}

func init() {
	installCmd.Flags().BoolVar(&installOverwrite, "overwrite", false, "overwrite existing files")
	installCmd.Flags().BoolVar(&installNonInteractive, "no-tui", false, "skip bubbletea TUI; use flag values")
	installCmd.Flags().StringVar(&installModeFlag, "mode", "minimal", "default verbosity mode")
	installCmd.Flags().StringVar(&installGuardFlag, "guard", "strict", "default guard level")
	installCmd.Flags().BoolVar(&installLockdownFlag, "lockdown", false, "prevent projects/env from loosening guard level")
	rootCmd.AddCommand(installCmd)
}
