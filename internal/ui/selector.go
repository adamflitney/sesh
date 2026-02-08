package ui

import (
	"fmt"
	"strings"

	"github.com/adamflitney/sesh/internal/finder"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

type model struct {
	projects  []finder.Project
	filtered  []finder.Project
	cursor    int
	textInput textinput.Model
	selected  *finder.Project
	quitting  bool
	err       error
	height    int
}

func initialModel(projects []finder.Project) model {
	ti := textinput.New()
	ti.Placeholder = "Search projects..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	return model{
		projects:  projects,
		filtered:  projects,
		cursor:    0,
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				m.selected = &m.filtered[m.cursor]
				m.quitting = true
				return m, tea.Quit
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}

		default:
			// Update text input
			m.textInput, cmd = m.textInput.Update(msg)

			// Filter projects based on search query
			query := m.textInput.Value()
			if query == "" {
				m.filtered = m.projects
			} else {
				m.filtered = m.fuzzyFilter(query)
			}

			// Reset cursor if it's out of bounds
			if m.cursor >= len(m.filtered) {
				m.cursor = 0
			}

			return m, cmd
		}
	}

	return m, cmd
}

func (m model) fuzzyFilter(query string) []finder.Project {
	var matches []finder.Project

	// Create a slice of project names for fuzzy matching
	names := make([]string, len(m.projects))
	for i, p := range m.projects {
		names[i] = p.Name
	}

	// Perform fuzzy search
	results := fuzzy.Find(query, names)

	// Build filtered list maintaining original project data
	for _, result := range results {
		matches = append(matches, m.projects[result.Index])
	}

	return matches
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("Select a project"))
	s.WriteString("\n\n")

	// Search input
	s.WriteString(m.textInput.View())
	s.WriteString("\n\n")

	// Error message if no projects
	if len(m.projects) == 0 {
		s.WriteString(errorStyle.Render("No Git projects found!"))
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("Make sure you have Git projects in your configured directories."))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press Esc or Ctrl+C to quit"))
		return s.String()
	}

	// No matches message
	if len(m.filtered) == 0 {
		s.WriteString(errorStyle.Render("No matches found"))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("↑/k up • ↓/j down • enter select • esc quit"))
		return s.String()
	}

	// Calculate how many items we can show
	maxItems := m.height - 10 // Account for header, input, and help text
	if maxItems < 5 {
		maxItems = 5
	}

	// Calculate visible range
	start := 0
	end := len(m.filtered)

	if len(m.filtered) > maxItems {
		// Center the cursor in the view
		start = m.cursor - maxItems/2
		if start < 0 {
			start = 0
		}
		end = start + maxItems
		if end > len(m.filtered) {
			end = len(m.filtered)
			start = end - maxItems
			if start < 0 {
				start = 0
			}
		}
	}

	// Show indicator if there are more items above
	if start > 0 {
		s.WriteString(helpStyle.Render(fmt.Sprintf("... %d more above ...", start)))
		s.WriteString("\n")
	}

	// Project list
	for i := start; i < end; i++ {
		project := m.filtered[i]

		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		name := project.Name
		path := project.Path

		if i == m.cursor {
			s.WriteString(cursor + selectedStyle.Render(name))
			s.WriteString("\n")
			s.WriteString("  " + pathStyle.Render(path))
		} else {
			s.WriteString(cursor + normalStyle.Render(name))
			s.WriteString("\n")
			s.WriteString("  " + pathStyle.Render(path))
		}

		s.WriteString("\n")
	}

	// Show indicator if there are more items below
	if end < len(m.filtered) {
		s.WriteString(helpStyle.Render(fmt.Sprintf("... %d more below ...", len(m.filtered)-end)))
		s.WriteString("\n")
	}

	// Help text
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("↑/k up • ↓/j down • enter select • esc quit"))

	return s.String()
}

// SelectProject displays a TUI for selecting a project and returns the selected project
func SelectProject(projects []finder.Project) (*finder.Project, error) {
	p := tea.NewProgram(initialModel(projects))

	m, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("error running program: %w", err)
	}

	finalModel := m.(model)
	if finalModel.err != nil {
		return nil, finalModel.err
	}

	return finalModel.selected, nil
}
