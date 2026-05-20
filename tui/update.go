package tui

import (
	"mendix-pvm/config"
	"mendix-pvm/project"
	"mendix-pvm/search"
	"mendix-pvm/version"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case syncDoneMsg:
		if msg.err != nil {
			m.errMsg = "Sync failed: " + msg.err.Error()
		} else {
			m.apps = append([]config.App{}, m.cfg.Apps...)
			if m.appCursor >= len(m.apps) {
				m.appCursor = 0
			}
			m.errMsg = ""
		}
		m.screen = screenDualPanel
		return m, nil

	case branchesLoadedMsg:
		if msg.err != nil {
			m.errMsg = "Failed to load branches: " + msg.err.Error()
			m.screen = screenBranchAction
		} else {
			m.remoteBranches = msg.branches
			m.remoteBranchCursor = 0
			m.screen = screenRemoteBranchList
			m.errMsg = ""
		}
		return m, nil

	case branchOpDoneMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			if m.branchMode == branchModeCheckout {
				m.screen = screenRemoteBranchList
			} else {
				m.screen = screenBranchNameInput
				m.nameInput.Focus()
			}
			return m, nil
		}
		_ = project.Open(msg.destDir)
		action := "Checked out"
		if m.branchMode == branchModeCreate {
			action = "Created"
		}
		m.statusMsg = action + " " + filepath.Base(msg.destDir)
		m.errMsg = ""
		m.screen = screenDualPanel
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenLoading:
		if msg.String() == "q" || msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	case screenDualPanel:
		return m.updateDualPanel(msg)
	case screenBranchList:
		return m.updateBranchList(msg)
	case screenBranchAction:
		return m.updateBranchAction(msg)
	case screenRemoteBranchList:
		return m.updateRemoteBranchList(msg)
	case screenBranchNameInput:
		return m.updateBranchNameInput(msg)
	}
	return m, nil
}

func (m model) updateDualPanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searching {
		return m.updateSearchMode(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "s":
		m.searching = true
		m.savedAppCursor = m.appCursor
		m.savedVersionCursor = m.versionCursor
		m.appCursor = 0
		m.versionCursor = 0
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		return m, nil
	case "o":
		_ = config.Open(m.cfg)
		return m, nil
	case "l", "right":
		if m.activePanel == panelApps {
			m.activePanel = panelVersions
		}
		return m, nil
	case "h", "left":
		if m.activePanel == panelVersions {
			m.activePanel = panelApps
		}
		return m, nil
	}

	if m.activePanel == panelApps {
		switch msg.String() {
		case "j", "down":
			if m.appCursor < len(m.apps)-1 {
				m.appCursor++
			}
		case "k", "up":
			if m.appCursor > 0 {
				m.appCursor--
			}
		case "enter":
			if len(m.apps) == 0 {
				return m, nil
			}
			app := m.apps[m.appCursor]
			m.selectedApp = app
			m.branches, _ = project.Search(m.cfg.ProjectDirectory, strings.Fields(app.Name))
			m.branchCursor = 0
			m.errMsg = ""
			m.statusMsg = ""
			m.screen = screenBranchList
		case "c":
			if len(m.apps) == 0 {
				return m, nil
			}
			m.selectedApp = m.apps[m.appCursor]
			m.branchActionCursor = 0
			m.branchActionOrigin = screenDualPanel
			m.errMsg = ""
			m.screen = screenBranchAction
		}
	} else {
		switch msg.String() {
		case "j", "down":
			if m.versionCursor < len(m.versions)-1 {
				m.versionCursor++
			}
		case "k", "up":
			if m.versionCursor > 0 {
				m.versionCursor--
			}
		case "enter":
			if len(m.versions) == 0 {
				return m, nil
			}
			vPath := m.versions[m.versionCursor]
			_ = version.Open(vPath)
			m.statusMsg = "Opened " + filepath.Base(vPath)
			m.errMsg = ""
		}
	}

	return m, nil
}

func (m model) updateSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEsc:
		m.searching = false
		m.appCursor = m.savedAppCursor
		m.versionCursor = m.savedVersionCursor
		m.searchInput.Blur()
		m.searchInput.SetValue("")
		return m, nil

	case tea.KeyEnter:
		query := m.searchInput.Value()
		filteredApps := search.SearchApps(m.apps, query)
		filteredVersions := search.FilterPaths(m.versions, query)
		m.searching = false
		m.searchInput.Blur()
		if m.activePanel == panelApps && len(filteredApps) > 0 {
			app := filteredApps[m.appCursor]
			m.selectedApp = app
			m.branches, _ = project.Search(m.cfg.ProjectDirectory, strings.Fields(app.Name))
			m.branchCursor = 0
			m.errMsg = ""
			m.screen = screenBranchList
		} else if m.activePanel == panelVersions && len(filteredVersions) > 0 {
			vPath := filteredVersions[m.versionCursor]
			_ = version.Open(vPath)
			m.statusMsg = "Opened " + filepath.Base(vPath)
			m.errMsg = ""
		}
		return m, nil

	case tea.KeyUp:
		if m.activePanel == panelApps {
			if m.appCursor > 0 {
				m.appCursor--
			}
		} else {
			if m.versionCursor > 0 {
				m.versionCursor--
			}
		}
		return m, nil

	case tea.KeyDown:
		query := m.searchInput.Value()
		if m.activePanel == panelApps {
			filtered := search.SearchApps(m.apps, query)
			if m.appCursor < len(filtered)-1 {
				m.appCursor++
			}
		} else {
			filtered := search.FilterPaths(m.versions, query)
			if m.versionCursor < len(filtered)-1 {
				m.versionCursor++
			}
		}
		return m, nil

	case tea.KeyLeft:
		m.activePanel = panelApps
		m.appCursor = 0
		return m, nil

	case tea.KeyRight:
		m.activePanel = panelVersions
		m.versionCursor = 0
		return m, nil
	}

	// All other keys (typing) go to the search input
	oldQuery := m.searchInput.Value()
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	newQuery := m.searchInput.Value()

	if oldQuery != newQuery {
		// Query changed — clamp cursors to the new filtered list sizes
		filteredApps := search.SearchApps(m.apps, newQuery)
		filteredVersions := search.FilterPaths(m.versions, newQuery)
		if m.appCursor >= len(filteredApps) {
			m.appCursor = max(0, len(filteredApps)-1)
		}
		if m.versionCursor >= len(filteredVersions) {
			m.versionCursor = max(0, len(filteredVersions)-1)
		}
	}

	return m, cmd
}

func (m model) updateBranchList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searching {
		return m.updateBranchSearchMode(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.screen = screenDualPanel
		m.errMsg = ""
	case "s":
		m.searching = true
		m.savedBranchCursor = m.branchCursor
		m.branchCursor = 0
		m.searchInput.SetValue("")
		m.searchInput.Focus()
	case "j", "down":
		if m.branchCursor < len(m.branches)-1 {
			m.branchCursor++
		}
	case "k", "up":
		if m.branchCursor > 0 {
			m.branchCursor--
		}
	case "enter":
		if len(m.branches) == 0 {
			return m, nil
		}
		branchPath := m.branches[m.branchCursor]
		_ = project.Open(branchPath)
		m.statusMsg = "Opened " + filepath.Base(branchPath)
		m.errMsg = ""
	case "c":
		m.branchActionCursor = 0
		m.branchActionOrigin = screenBranchList
		m.errMsg = ""
		m.screen = screenBranchAction
	}
	return m, nil
}

func (m model) updateBranchSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEsc:
		m.searching = false
		m.branchCursor = m.savedBranchCursor
		m.searchInput.Blur()
		m.searchInput.SetValue("")
		return m, nil

	case tea.KeyEnter:
		filtered := search.FilterPaths(m.branches, m.searchInput.Value())
		m.searching = false
		m.searchInput.Blur()
		if len(filtered) > 0 {
			branchPath := filtered[m.branchCursor]
			_ = project.Open(branchPath)
			m.statusMsg = "Opened " + filepath.Base(branchPath)
			m.errMsg = ""
		}
		return m, nil

	case tea.KeyUp:
		if m.branchCursor > 0 {
			m.branchCursor--
		}
		return m, nil

	case tea.KeyDown:
		filtered := search.FilterPaths(m.branches, m.searchInput.Value())
		if m.branchCursor < len(filtered)-1 {
			m.branchCursor++
		}
		return m, nil
	}

	oldQuery := m.searchInput.Value()
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	newQuery := m.searchInput.Value()

	if oldQuery != newQuery {
		filtered := search.FilterPaths(m.branches, newQuery)
		if m.branchCursor >= len(filtered) {
			m.branchCursor = max(0, len(filtered)-1)
		}
	}

	return m, cmd
}

func (m model) updateBranchAction(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.screen = m.branchActionOrigin
		m.errMsg = ""
	case "j", "down":
		if m.branchActionCursor < 1 {
			m.branchActionCursor++
		}
	case "k", "up":
		if m.branchActionCursor > 0 {
			m.branchActionCursor--
		}
	case "enter":
		if m.branchActionCursor == 0 {
			m.branchMode = branchModeCreate
		} else {
			m.branchMode = branchModeCheckout
		}
		if m.pat == "" {
			m.errMsg = "MX_PAT not set. Run 'mx config' to configure it."
			return m, nil
		}
		if m.selectedApp.AppID == "" {
			m.errMsg = "App ID not found. Run 'mx sync' to refresh the app list."
			return m, nil
		}
		m.screen = screenLoading
		m.loadingMsg = "Loading remote branches..."
		m.errMsg = ""
		return m, getBranchesCmd(m.ctx, m.pat, m.selectedApp.AppID)
	}
	return m, nil
}

func (m model) updateRemoteBranchList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.screen = screenBranchAction
		m.errMsg = ""
	case "j", "down":
		if m.remoteBranchCursor < len(m.remoteBranches)-1 {
			m.remoteBranchCursor++
		}
	case "k", "up":
		if m.remoteBranchCursor > 0 {
			m.remoteBranchCursor--
		}
	case "enter":
		if len(m.remoteBranches) == 0 {
			return m, nil
		}
		selected := m.remoteBranches[m.remoteBranchCursor].Name
		if m.branchMode == branchModeCheckout {
			m.screen = screenLoading
			m.loadingMsg = "Checking out '" + selected + "'..."
			m.errMsg = ""
			return m, checkoutBranchCmd(m.ctx, m.cfg, m.selectedApp, selected)
		}
		m.selectedBase = selected
		m.screen = screenBranchNameInput
		m.nameInput.SetValue("")
		m.nameInput.Focus()
		m.errMsg = ""
	}
	return m, nil
}

func (m model) updateBranchNameInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.screen = screenRemoteBranchList
		m.nameInput.Blur()
		m.errMsg = ""
		return m, nil
	case tea.KeyEnter:
		branchName := strings.TrimSpace(m.nameInput.Value())
		if branchName == "" {
			m.errMsg = "Branch name cannot be empty."
			return m, nil
		}
		m.screen = screenLoading
		m.loadingMsg = "Creating branch '" + branchName + "'..."
		m.errMsg = ""
		return m, createBranchCmd(m.ctx, m.cfg, m.selectedApp, branchName, m.selectedBase)
	}
	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}
