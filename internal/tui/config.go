package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// ConfigModel holds the state for our config TUI
type ConfigModel struct {
	version      string
	inContainer  bool
	dockerExists bool
	step         int
	totalSteps   int
	selectedMode int // 0 = quick, 1 = advanced
	inputs       []textinput.Model
	choices      map[int]string // user choices
	finished     bool
	err          error
}

// NewConfigModel creates a new config TUI model
func NewConfigModel(version string, inContainer bool, dockerExists bool) ConfigModel {
	inputs := make([]textinput.Model, 4)

	// Registry mirror
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "registry-1.docker.io"
	inputs[0].Prompt = "Registry mirror: "
	inputs[0].Width = 50

	// HTTP proxy
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "http://proxy.example.com:8080"
	inputs[1].Prompt = "HTTP proxy: "
	inputs[1].Width = 50

	// HTTPS proxy
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "http://proxy.example.com:8080"
	inputs[2].Prompt = "HTTPS proxy: "
	inputs[2].Width = 50

	// Container CIDR
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "172.31.0.0/16"
	inputs[3].Prompt = "Container network CIDR: "
	inputs[3].Width = 50

	inputs[0].Focus()

	return ConfigModel{
		version:      version,
		inContainer:  inContainer,
		dockerExists: dockerExists,
		step:         0,
		totalSteps:   2, // default to quick mode
		selectedMode: 0,
		inputs:       inputs,
		choices:      make(map[int]string),
	}
}

// Init initializes the model
func (m ConfigModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the model
func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			if m.step == 0 {
				// Mode selection
				m.step++
				if m.selectedMode == 1 {
					m.totalSteps = 5
				}
			} else if m.step == 1 {
				// System check -> install
				if m.selectedMode == 0 {
					// Quick mode, done with TUI
					m.finished = true
					return m, tea.Quit
				} else {
					// Advanced mode, continue
					m.step++
				}
			} else if m.step >= 2 && m.step <= 4 {
				// Input steps for advanced mode
				if m.step == 2 {
					// Registry
					m.choices[0] = m.inputs[0].Value()
					m.step++
					m.inputs[1].Focus()
				} else if m.step == 3 {
					// Proxy
					m.choices[1] = m.inputs[1].Value()
					m.choices[2] = m.inputs[2].Value()
					m.step++
					m.inputs[3].Focus()
				} else if m.step == 4 {
					// CIDR
					m.choices[3] = m.inputs[3].Value()
					m.finished = true
					return m, tea.Quit
				}
			}

		case "up", "k":
			if m.step == 0 && m.selectedMode > 0 {
				m.selectedMode--
			}

		case "down", "j":
			if m.step == 0 && m.selectedMode < 1 {
				m.selectedMode++
			}

		case "tab":
			if m.step >= 2 && m.step <= 4 && m.selectedMode == 1 {
				idx := m.step - 2
				m.inputs[idx].Blur()
				idx = (idx + 1) % 4
				m.inputs[idx].Focus()
			}
		}
	}

	// Update text inputs
	var cmd tea.Cmd
	if m.step >= 2 && m.step <= 4 && m.selectedMode == 1 {
		if m.step == 2 {
			m.inputs[0], cmd = m.inputs[0].Update(msg)
		} else if m.step == 3 {
			m.inputs[1], cmd = m.inputs[1].Update(msg)
			m.inputs[2], _ = m.inputs[2].Update(msg)
		} else if m.step == 4 {
			m.inputs[3], cmd = m.inputs[3].Update(msg)
		}
	}

	return m, cmd
}

// View renders the UI
func (m ConfigModel) View() string {
	var s strings.Builder

	// Banner
	s.WriteString(m.renderBanner())
	s.WriteString("\n")

	// Progress bar
	s.WriteString(m.renderProgress())
	s.WriteString("\n\n")

	// Current step content
	switch m.step {
	case 0:
		s.WriteString(m.renderModeSelection())
	case 1:
		s.WriteString(m.renderSystemCheck())
	case 2:
		s.WriteString(m.renderRegistryConfig())
	case 3:
		s.WriteString(m.renderProxyConfig())
	case 4:
		s.WriteString(m.renderCIDRConfig())
	}

	// Help footer
	s.WriteString("\n\n")
	s.WriteString(m.renderHelp())

	return s.String()
}

func (m ConfigModel) renderBanner() string {
	bannerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true)

	return bannerStyle.Render(fmt.Sprintf(
		`═══════════════════════════════════════════════════════════
  Docker Pilot  |  Version: %s
  Docker setup & TUI management for SLES 15+
═══════════════════════════════════════════════════════════`,
		m.version,
	))
}

func (m ConfigModel) renderProgress() string {
	progress := fmt.Sprintf("[%d/%d]", m.step+1, m.totalSteps)
	stepTitles := []string{
		"Select Mode",
		"System Check & Install",
		"Configure Registry",
		"Configure Proxy",
		"Configure Network",
	}

	title := ""
	if m.step < len(stepTitles) {
		title = stepTitles[m.step]
	}

	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700"))

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		progressStyle.Render(progress),
		" ",
		titleStyle.Render(title),
	)
}

func (m ConfigModel) renderModeSelection() string {
	var s strings.Builder

	options := []string{
		"Quick mode (install Docker only, skip configuration)",
		"Advanced mode (full setup: install + configure registry/proxy/network)",
	}

	s.WriteString("Please select installation mode:\n\n")

	for i, option := range options {
		cursor := " "
		if i == m.selectedMode {
			cursor = ">"
		}
		if i == m.selectedMode {
			selectedStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)
			s.WriteString(selectedStyle.Render(fmt.Sprintf("%s %s\n", cursor, option)))
		} else {
			s.WriteString(fmt.Sprintf("%s %s\n", cursor, option))
		}
	}

	return s.String()
}

func (m ConfigModel) renderSystemCheck() string {
	var s strings.Builder

	s.WriteString("System check...\n")

	// In container warning
	if m.inContainer {
		warningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00"))
		s.WriteString(warningStyle.Render("⚠️  Running in container - Docker install skipped\n"))
	}

	// Docker exists
	if m.dockerExists && !m.inContainer {
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF"))
		s.WriteString(infoStyle.Render("ℹ️  Docker is already installed\n"))
	}

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))
	s.WriteString(successStyle.Render("✓ System check passed\n"))

	s.WriteString("\nPress Enter to continue...")

	return s.String()
}

func (m ConfigModel) renderRegistryConfig() string {
	var s strings.Builder

	s.WriteString("Configure registry mirror (optional)\n")
	s.WriteString("Press Enter to keep default or skip\n\n")
	s.WriteString(m.inputs[0].View())

	return s.String()
}

func (m ConfigModel) renderProxyConfig() string {
	var s strings.Builder

	s.WriteString("Configure HTTP/HTTPS proxy (optional)\n")
	s.WriteString("Fill in either or both, then press Enter to continue\n\n")
	s.WriteString("HTTP Proxy:\n")
	s.WriteString(m.inputs[1].View())
	s.WriteString("\n\nHTTPS Proxy:\n")
	s.WriteString(m.inputs[2].View())
	s.WriteString("\n\nUse Tab to switch between inputs")

	return s.String()
}

func (m ConfigModel) renderCIDRConfig() string {
	var s strings.Builder

	s.WriteString("Configure container network CIDR (optional)\n")
	s.WriteString("Avoid conflicts with internal networks\n\n")
	s.WriteString(m.inputs[3].View())

	return s.String()
}

func (m ConfigModel) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	return helpStyle.Render("↑/↓ or k/j: Navigate • Enter: Select • Ctrl+C: Quit")
}

// Finished returns whether the config flow is complete
func (m ConfigModel) Finished() bool {
	return m.finished
}

// IsAdvancedMode returns whether user selected advanced mode
func (m ConfigModel) IsAdvancedMode() bool {
	return m.selectedMode == 1
}

// GetChoice returns the user's choice for a specific input index
func (m ConfigModel) GetChoice(idx int) string {
	return m.choices[idx]
}
