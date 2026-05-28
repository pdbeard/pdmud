package main

import (
	"bufio"
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

var tabs = []struct {
	label   string
	logFile string
}{
	{"Global", "logs/global.log"},
	{"Local", "logs/local.log"},
	{"Tell", "logs/tells.log"},
	{"Auction", "logs/auction.log"},
}

const maxLines = 200

// ---------------------------------------------------------------
// Styles
// ---------------------------------------------------------------

var (
	tabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Padding(0, 1)

	tabInactive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	tabBar = lipgloss.NewStyle().
		Background(lipgloss.Color("236"))

	contentStyle = lipgloss.NewStyle().
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("57"))
)

// ---------------------------------------------------------------
// Messages
// ---------------------------------------------------------------

type newLineMsg struct {
	tabIdx int
	line   string
}

type tickMsg time.Time

// ---------------------------------------------------------------
// Model
// ---------------------------------------------------------------

type model struct {
	activeTab int
	lines     [4][]string
	width     int
	height    int
	// tab label pixel offsets for click detection
	tabOffsets []int
}

func initialModel() model {
	return model{activeTab: 0}
}

// ---------------------------------------------------------------
// Init
// ---------------------------------------------------------------

func (m model) Init() tea.Cmd {
	cmds := make([]tea.Cmd, len(tabs))
	for i, t := range tabs {
		cmds[i] = watchFile(i, t.logFile)
	}
	return tea.Batch(cmds...)
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
			if msg.Y == 1 { // tab bar row (inside border)
				m.activeTab = tabAtX(m.tabOffsets, msg.X)
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1", "2", "3", "4":
			idx := int(msg.String()[0]-'1')
			if idx < len(tabs) {
				m.activeTab = idx
			}
		case "tab":
			m.activeTab = (m.activeTab + 1) % len(tabs)
		}

	case newLineMsg:
		m.lines[msg.tabIdx] = appendLine(m.lines[msg.tabIdx], msg.line)
		return m, watchFile(msg.tabIdx, tabs[msg.tabIdx].logFile)
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

	inner := m.width - 2   // account for border
	contentH := m.height - 4 // border(2) + tabbar(1) + divider(1)
	if contentH < 1 {
		contentH = 1
	}

	// Tab bar
	bar := renderTabBar(m.activeTab, inner)

	// Content — show last N lines that fit
	lines := m.lines[m.activeTab]
	visible := lines
	if len(visible) > contentH {
		visible = visible[len(visible)-contentH:]
	}
	// Pad to fill height
	for len(visible) < contentH {
		visible = append(visible, "")
	}
	content := contentStyle.Width(inner).Render(strings.Join(visible, "\n"))

	body := lipgloss.JoinVertical(lipgloss.Left, bar, content)
	return borderStyle.Width(m.width - 2).Render(body)
}

// ---------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------

func renderTabBar(active, width int) string {
	var parts []string
	for i, t := range tabs {
		if i == active {
			parts = append(parts, tabActive.Render(t.label))
		} else {
			parts = append(parts, tabInactive.Render(t.label))
		}
	}
	bar := strings.Join(parts, " ")
	return tabBar.Width(width).Render(bar)
}

func buildTabOffsets() []int {
	offsets := make([]int, len(tabs))
	x := 1 // start inside border
	for i, t := range tabs {
		offsets[i] = x
		x += len(t.label) + 3 // padding(2) + space(1)
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

func appendLine(lines []string, line string) []string {
	lines = append(lines, line)
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return lines
}

// watchFile tails a log file, blocking until a new line appears,
// then returns a newLineMsg. Called repeatedly via Cmd chaining.
func watchFile(tabIdx int, path string) tea.Cmd {
	return func() tea.Msg {
		abs, _ := filepath.Abs(path)
		f, err := os.Open(abs)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			return newLineMsg{tabIdx: tabIdx, line: ""}
		}
		defer f.Close()

		// Seek to end so we only get new lines
		f.Seek(0, 2)
		scanner := bufio.NewScanner(f)

		for {
			for scanner.Scan() {
				line := scanner.Text()
				if line != "" {
					return newLineMsg{tabIdx: tabIdx, line: line}
				}
			}
			time.Sleep(100 * time.Millisecond)
			// Reset scanner after EOF
			f.Seek(0, 2)
			scanner = bufio.NewScanner(f)
		}
	}
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
