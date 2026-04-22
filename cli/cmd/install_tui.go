package cmd

import (
	"fmt"
	"io/fs"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type installModel struct {
	claudeDir string
	step      int
	mode      string
	guard     string
	lockdown  bool
	confirmed bool
	aborted   bool
}

func newInstallModel(claudeDir string) installModel {
	return installModel{claudeDir: claudeDir, mode: "minimal", guard: "strict"}
}

func (m installModel) Init() tea.Cmd { return nil }

func (m installModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch k.String() {
	case "ctrl+c", "esc":
		m.aborted = true
		return m, tea.Quit
	case "enter":
		m.step++
		if m.step > 3 {
			m.confirmed = true
			return m, tea.Quit
		}
		return m, nil
	case "left", "h":
		m = cycleValue(m, -1)
		return m, nil
	case "right", "l":
		m = cycleValue(m, 1)
		return m, nil
	}
	return m, nil
}

func cycleValue(m installModel, dir int) installModel {
	switch m.step {
	case 0:
		modes := []string{"silent", "minimal", "normal", "verbose", "caveman"}
		m.mode = rotate(modes, m.mode, dir)
	case 1:
		levels := []string{"strict", "project", "open", "off"}
		m.guard = rotate(levels, m.guard, dir)
	case 2:
		m.lockdown = !m.lockdown
	}
	return m
}

func rotate(values []string, cur string, dir int) string {
	idx := 0
	for i, v := range values {
		if v == cur {
			idx = i
			break
		}
	}
	idx = (idx + dir + len(values)) % len(values)
	return values[idx]
}

var titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))

func (m installModel) View() string {
	b := &strings.Builder{}
	fmt.Fprintln(b, titleStyle.Render("IFLy install"))
	fmt.Fprintf(b, "claude dir: %s\n\n", m.claudeDir)
	fmt.Fprintf(b, "  [%s] mode:     %s\n", mark(m.step == 0), m.mode)
	fmt.Fprintf(b, "  [%s] guard:    %s\n", mark(m.step == 1), m.guard)
	fmt.Fprintf(b, "  [%s] lockdown: %v\n", mark(m.step == 2), m.lockdown)
	fmt.Fprintf(b, "  [%s] confirm\n", mark(m.step == 3))
	fmt.Fprintln(b, "\nenter: next   \u2190 \u2192 : change   esc: cancel")
	return b.String()
}

func mark(active bool) string {
	if active {
		return "*"
	}
	return " "
}

func (m installModel) opts(pluginFS fs.FS, overwrite bool) InstallOpts {
	return InstallOpts{
		PluginFS:  pluginFS,
		ClaudeDir: m.claudeDir,
		Mode:      m.mode,
		Guard:     m.guard,
		Lockdown:  m.lockdown,
		Overwrite: overwrite,
	}
}
