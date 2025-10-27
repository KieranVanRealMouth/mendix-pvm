package ui

import (
	"strings"

	"mendix-project-manager/internal/models"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type TableModel struct {
	versionTable table.Model
	projectTable table.Model
	activeTable  int // 0 for versions, 1 for projects
	choice       *models.SearchResult
	quitting     bool
	hasVersions  bool
	hasProjects  bool
}

func NewTableModel(versions []models.Version, projects []models.Project) TableModel {
	var versionTable, projectTable table.Model
	hasVersions := len(versions) > 0
	hasProjects := len(projects) > 0

	if hasVersions {
		versionColumns := []table.Column{
			{Title: "Version", Width: 20},
			{Title: "Path", Width: 60},
		}

		versionRows := make([]table.Row, len(versions))
		for i, v := range versions {
			versionRows[i] = table.Row{v.String(), v.Path}
		}

		versionTable = table.New(
			table.WithColumns(versionColumns),
			table.WithRows(versionRows),
			table.WithFocused(true),
			table.WithHeight(10),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		versionTable.SetStyles(s)
	}

	if hasProjects {
		projectColumns := []table.Column{
			{Title: "Project", Width: 30},
			{Title: "Path", Width: 50},
		}

		projectRows := make([]table.Row, len(projects))
		for i, p := range projects {
			projectRows[i] = table.Row{p.Name, p.Path}
		}

		projectTable = table.New(
			table.WithColumns(projectColumns),
			table.WithRows(projectRows),
			table.WithFocused(!hasVersions),
			table.WithHeight(10),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		projectTable.SetStyles(s)
	}

	activeTable := 0
	if !hasVersions && hasProjects {
		activeTable = 1
	}

	return TableModel{
		versionTable: versionTable,
		projectTable: projectTable,
		activeTable:  activeTable,
		hasVersions:  hasVersions,
		hasProjects:  hasProjects,
	}
}

func (m TableModel) Init() tea.Cmd {
	return nil
}

func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.activeTable == 0 && m.hasVersions {
				// Version selected
				// TODO: Require project to be selected
			} else if m.activeTable == 1 && m.hasProjects {
				// TODO: Project selected
				// TODO: Implement opening project with a version when both are opened
				// TODO: Implement description of usage of TUI
			}
			return m, tea.Quit
		case "tab":
			// TODO: Fix the table switch to be done using arrow left and right
			if m.hasVersions && m.hasProjects {
				if m.activeTable == 0 {
					m.activeTable = 1
					m.versionTable.Blur()
					m.projectTable.Focus()
				} else {
					m.activeTable = 0
					m.projectTable.Blur()
					m.versionTable.Focus()
				}
			}
		}
	}

	if m.activeTable == 0 && m.hasVersions {
		m.versionTable, cmd = m.versionTable.Update(msg)
	} else if m.activeTable == 1 && m.hasProjects {
		m.projectTable, cmd = m.projectTable.Update(msg)
	}

	return m, cmd
}

func (m TableModel) View() string {
	var b strings.Builder

	if m.hasVersions {
		b.WriteString("Versions:\n")
		b.WriteString(baseStyle.Render(m.versionTable.View()))
		b.WriteString("\n\n")
	}

	if m.hasProjects {
		b.WriteString("Projects:\n")
		b.WriteString(baseStyle.Render(m.projectTable.View()))
		b.WriteString("\n")
	}

	if m.hasVersions && m.hasProjects {
		b.WriteString("\nPress Tab to switch between tables, Enter to select, q to quit\n")
	} else {
		b.WriteString("\nPress Enter to select, q to quit\n")
	}

	return b.String()
}
