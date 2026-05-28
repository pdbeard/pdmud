package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------
// Config
// ---------------------------------------------------------------

const (
	stateAffects = "logs/affects.state"
	stateParty   = "logs/party.state"
	pollInterval = 500 * time.Millisecond
)

var tabLabels = []string{"Affects", "Party"}

// ---------------------------------------------------------------
// Styles
// ---------------------------------------------------------------

var (
	tabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("28")).
			Padding(0, 1)

	tabInactive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	tabBar = lipgloss.NewStyle().
		Background(lipgloss.Color("236"))

	affectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("114")).
			Padding(0, 1)

	partyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Width(10)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("28"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Padding(0, 1)
)

// ---------------------------------------------------------------
// Messages
// ---------------------------------------------------------------

type pollMsg struct{}

// ---------------------------------------------------------------
// Model
// ---------------------------------------------------------------

type model struct {
	activeTab  int
	affects    []string
	party      []string
	width      int
	height     int
	tabOffsets []int
}

func initialModel() model {
	return model{activeTab: 0}
}

// ---------------------------------------------------------------
// Init
// ---------------------------------------------------------------

func (m model) Init() tea.Cmd {
	return poll()
}

func poll() tea.Cmd {
	return tea.Tick(pollInterval, func(t time.Time) tea.Msg {
		return pollMsg{}
	})
}

// ---------------------------------------------------------------
// Update
// ---------------------------------------------------------------

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tabOffsets = buildTabOffsets()

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if msg.Y == 1 {
				m.activeTab = tabAtX(m.tabOffsets, msg.X)
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % len(tabLabels)
		case "1", "2":
			idx := int(msg.String()[0] - '1')
			if idx < len(tabLabels) {
				m.activeTab = idx
			}
		}

	case pollMsg:
		m.affects = readStateFile(stateAffects)
		m.party = readStateFile(stateParty)
		return m, poll()
	}

	return m, nil
}

// ---------------------------------------------------------------
// View
// ---------------------------------------------------------------

func (m model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	inner := m.width - 2
	contentH := m.height - 4
	if contentH < 1 {
		contentH = 1
	}

	bar := renderTabBar(m.activeTab, inner)

	var content string
	switch m.activeTab {
	case 0:
		content = renderAffects(m.affects, inner, contentH)
	case 1:
		content = renderParty(m.party, inner, contentH)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, bar, content)
	return borderStyle.Width(m.width - 2).Render(body)
}

// ---------------------------------------------------------------
// Renderers
// ---------------------------------------------------------------

func renderAffects(affects []string, width, height int) string {
	if len(affects) == 0 {
		return emptyStyle.Render("No active affects")
	}
	lines := affects
	if len(lines) > height {
		lines = lines[:height]
	}
	rows := make([]string, len(lines))
	for i, a := range lines {
		rows[i] = affectStyle.Width(width).Render(a)
	}
	return strings.Join(rows, "\n")
}

func renderParty(party []string, width, height int) string {
	if len(party) == 0 {
		return emptyStyle.Render("Not in a party")
	}
	lines := party
	if len(lines) > height {
		lines = lines[:height]
	}
	rows := make([]string, len(lines))
	for i, p := range lines {
		parts := strings.SplitN(p, " ", 2)
		name := ""
		rest := p
		if len(parts) == 2 {
			name = parts[0]
			rest = parts[1]
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render(name),
			partyStyle.Render(rest),
		)
		rows[i] = row
	}
	return strings.Join(rows, "\n")
}

// ---------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------

func renderTabBar(active, width int) string {
	var parts []string
	for i, label := range tabLabels {
		if i == active {
			parts = append(parts, tabActive.Render(label))
		} else {
			parts = append(parts, tabInactive.Render(label))
		}
	}
	return tabBar.Width(width).Render(strings.Join(parts, " "))
}

func buildTabOffsets() []int {
	offsets := make([]int, len(tabLabels))
	x := 1
	for i, label := range tabLabels {
		offsets[i] = x
		x += len(label) + 3
	}
	return offsets
}

func tabAtX(offsets []int, x int) int {
	active := 0
	for i, offset := range offsets {
		if x >= offset {
			active = i
		}
	}
	return active
}

func readStateFile(path string) []string {
	abs, _ := filepath.Abs(path)
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil
	}
	var lines []string
	for _, l := range strings.Split(string(data), "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

// ---------------------------------------------------------------
// Main
// ---------------------------------------------------------------

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
