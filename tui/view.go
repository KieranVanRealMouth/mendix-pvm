package tui

import (
	"fmt"
	"mendix-pvm/search"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	switch m.screen {
	case screenLoading:
		return m.viewLoading()
	case screenDualPanel:
		return m.viewDualPanel()
	case screenBranchList:
		return m.viewBranchList()
	case screenBranchAction:
		return m.viewBranchAction()
	case screenRemoteBranchList:
		return m.viewRemoteBranchList()
	case screenBranchNameInput:
		return m.viewBranchNameInput()
	}
	return ""
}

func (m model) viewLoading() string {
	return "\n  " + m.spinner.View() + " " + m.loadingMsg + "\n"
}

func (m model) viewDualPanel() string {
	lw := m.width / 2
	rw := m.width - lw

	// Reserve one extra line when the search bar is visible
	reserved := 4
	if m.searching {
		reserved++
	}
	listH := m.height - reserved
	if listH < 1 {
		listH = 1
	}

	// Compute displayed lists (filtered when searching, full otherwise)
	query := m.searchInput.Value()
	displayApps := search.SearchApps(m.apps, query)
	displayVersions := search.FilterPaths(m.versions, query)
	if !m.searching || query == "" {
		displayApps = m.apps
		displayVersions = m.versions
	}

	// Clamp cursors for rendering safety
	appCursor := m.appCursor
	if appCursor >= len(displayApps) {
		appCursor = max(0, len(displayApps)-1)
	}
	versionCursor := m.versionCursor
	if versionCursor >= len(displayVersions) {
		versionCursor = max(0, len(displayVersions)-1)
	}

	// Left panel: Apps
	var leftHeader string
	if m.activePanel == panelApps {
		leftHeader = activeHeaderStyle.Render("Apps")
	} else {
		leftHeader = inactiveHeaderStyle.Render("Apps")
	}
	appNames := make([]string, len(displayApps))
	for i, a := range displayApps {
		appNames[i] = a.Name
	}
	leftContent := leftHeader + "\n" + strings.Repeat("─", lw) + "\n" +
		renderItemList(appNames, appCursor, listH, lw-2)

	// Right panel: Versions
	var rightHeader string
	if m.activePanel == panelVersions {
		rightHeader = activeHeaderStyle.Render("Versions")
	} else {
		rightHeader = inactiveHeaderStyle.Render("Versions")
	}
	versionNames := make([]string, len(displayVersions))
	for i, v := range displayVersions {
		versionNames[i] = filepath.Base(v)
	}
	rightContent := rightHeader + "\n" + strings.Repeat("─", rw) + "\n" +
		renderItemList(versionNames, versionCursor, listH, rw-2)

	panels := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(lw).Render(leftContent),
		lipgloss.NewStyle().Width(rw).Render(rightContent),
	)

	var searchBar string
	if m.searching {
		searchBar = "\n  Search: " + m.searchInput.View()
	}

	var footer string
	if m.searching {
		footer = footerStyle.Render("(↑/↓) navigate results  (←/→) switch panels  (enter) select  (esc) cancel")
	} else {
		footer = footerStyle.Render("(j/k) navigate  (l/→) versions  (h/←) apps  (enter) branches  (c) create/checkout  (s) search  (o) config  (q) quit")
	}

	var errLine string
	if m.errMsg != "" {
		errLine = "\n" + errorStyle.Render(m.errMsg)
	}

	return panels + searchBar + errLine + "\n" + footer
}

func (m model) viewBranchList() string {
	title := "Local branches — " + m.selectedApp.Name

	reserved := 5
	if m.searching {
		reserved++
	}
	listH := m.height - reserved
	if listH < 1 {
		listH = 1
	}

	displayBranches := m.branches
	if m.searching && m.searchInput.Value() != "" {
		displayBranches = search.FilterPaths(m.branches, m.searchInput.Value())
	}

	branchCursor := m.branchCursor
	if branchCursor >= len(displayBranches) {
		branchCursor = max(0, len(displayBranches)-1)
	}

	items := make([]string, len(displayBranches))
	for i, b := range displayBranches {
		items[i] = filepath.Base(b)
	}

	body := renderItemList(items, branchCursor, listH, m.width-3)
	if m.searching {
		body += "  Search: " + m.searchInput.View() + "\n"
	}

	var footer string
	if m.searching {
		footer = footerStyle.Render("(↑/↓) navigate results  (enter) open  (esc) cancel search")
	} else {
		footer = footerStyle.Render("(j/k) navigate  (enter) open  (c) create/checkout  (s) search  (esc) back  (q) quit")
	}
	return renderScreen(title, m.width, body, m.errMsg, footer)
}

func (m model) viewBranchAction() string {
	title := m.selectedApp.Name + " — what would you like to do?"
	items := []string{"Create branch", "Checkout branch"}
	listH := m.height - 5
	if listH < 1 {
		listH = 1
	}
	footer := footerStyle.Render("(j/k) navigate  (enter) select  (esc) back")
	return renderScreen(title, m.width, renderItemList(items, m.branchActionCursor, listH, m.width-3), m.errMsg, footer)
}

func (m model) viewRemoteBranchList() string {
	var title string
	if m.branchMode == branchModeCreate {
		title = "Select base branch — " + m.selectedApp.Name
	} else {
		title = "Select branch to checkout — " + m.selectedApp.Name
	}

	items := make([]string, len(m.remoteBranches))
	for i, b := range m.remoteBranches {
		items[i] = b.Name
	}
	listH := m.height - 5
	if listH < 1 {
		listH = 1
	}

	var footerText string
	if m.branchMode == branchModeCreate {
		footerText = "(j/k) navigate  (enter) use as base  (esc) back  (q) quit"
	} else {
		footerText = "(j/k) navigate  (enter) checkout  (esc) back  (q) quit"
	}
	footer := footerStyle.Render(footerText)
	return renderScreen(title, m.width, renderItemList(items, m.remoteBranchCursor, listH, m.width-3), m.errMsg, footer)
}

func (m model) viewBranchNameInput() string {
	title := fmt.Sprintf("New branch from '%s' — %s", m.selectedBase, m.selectedApp.Name)
	body := "\n  Branch name: " + m.nameInput.View() + "\n\n"
	footer := footerStyle.Render("(enter) create  (esc) cancel")
	return renderScreen(title, m.width, body, m.errMsg, footer)
}

// renderScreen renders a full-width screen with a title, separator, body, optional error, and footer.
func renderScreen(title string, width int, body, errMsg, footer string) string {
	sep := strings.Repeat("─", width)
	var errLine string
	if errMsg != "" {
		errLine = errorStyle.Render(errMsg) + "\n"
	}
	return title + "\n" + sep + "\n" + body + errLine + footer
}

// renderItemList renders a scrollable list of items with the cursor highlighted.
func renderItemList(items []string, cursor, maxH, itemWidth int) string {
	if len(items) == 0 {
		return footerStyle.Render("  (empty)") + "\n"
	}
	if itemWidth < 4 {
		itemWidth = 4
	}

	// Compute visible window centred on cursor
	start := 0
	if len(items) > maxH {
		start = cursor - maxH/2
		if start < 0 {
			start = 0
		}
		if start+maxH > len(items) {
			start = len(items) - maxH
		}
	}
	end := start + maxH
	if end > len(items) {
		end = len(items)
	}

	maxNameLen := itemWidth - 2
	var sb strings.Builder
	for i := start; i < end; i++ {
		name := items[i]
		if len(name) > maxNameLen && maxNameLen > 3 {
			name = name[:maxNameLen-3] + "..."
		}
		if i == cursor {
			sb.WriteString(selectedItemStyle.Width(itemWidth).Render("> " + name))
		} else {
			sb.WriteString("  " + name)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
