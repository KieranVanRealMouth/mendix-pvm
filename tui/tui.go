package tui

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mendix-pvm/branch"
	"mendix-pvm/config"
	"mendix-pvm/platform"
	"mendix-pvm/version"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenDualPanel screen = iota
	screenBranchList
	screenBranchAction
	screenRemoteBranchList
	screenBranchNameInput
	screenLoading
)

type activePanel int

const (
	panelApps     activePanel = iota
	panelVersions
)

type branchMode int

const (
	branchModeCreate   branchMode = iota
	branchModeCheckout
)

// Messages

type syncDoneMsg struct{ err error }

type branchesLoadedMsg struct {
	branches []platform.Branch
	err      error
}

type branchOpDoneMsg struct {
	jobID   int
	destDir string
	err     error
}

type clearJobMsg struct{ id int }

// bgJob tracks a background branch operation displayed in the footer.
type bgJob struct {
	id    int
	label string
	done  bool
	err   error
}

type model struct {
	cfg       *config.Config
	pat       string
	ctx       context.Context
	cancelCtx context.CancelFunc

	screen      screen
	activePanel activePanel

	// App list
	appCursor int
	apps      []config.App

	// Version list
	versionCursor int
	versions      []string

	// Selected app + local branch list
	selectedApp        config.App
	branchCursor       int
	branches           []string
	branchActionOrigin screen

	// Branch action picker (0=Create, 1=Checkout)
	branchActionCursor int
	branchMode         branchMode

	// Remote branch list
	remoteBranchCursor int
	remoteBranches     []platform.Branch
	selectedBase       string

	// Branch name input
	nameInput textinput.Model

	// Inline search (s key in dual panel, branch list, and remote branch list)
	searching                bool
	searchInput              textinput.Model
	savedAppCursor           int
	savedVersionCursor       int
	savedBranchCursor        int
	savedRemoteBranchCursor  int

	// Loading
	spinner    spinner.Model
	loadingMsg string

	// Background jobs shown in footer
	bgJobs        []bgJob
	nextJobID     int
	confirmingQuit bool

	// Inline error / status
	errMsg    string
	statusMsg string

	// Terminal size
	width  int
	height int
}

func Run(cfg *config.Config) error {
	versions, _ := version.Search(cfg.VersionDirectory, nil)
	pat := os.Getenv("MX_PAT")

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	ti := textinput.New()
	ti.Placeholder = "feat/my-new-feature"
	ti.CharLimit = 100

	si := textinput.New()
	si.Placeholder = "search apps and versions..."
	si.CharLimit = 60

	ctx, cancel := context.WithCancel(context.Background())
	m := model{
		cfg:       cfg,
		pat:       pat,
		ctx:       ctx,
		cancelCtx: cancel,
		apps:        append([]config.App{}, cfg.Apps...),
		versions:    versions,
		spinner:     sp,
		nameInput:   ti,
		searchInput: si,
		width:     80,
		height:    24,
	}

	if len(m.apps) == 0 {
		if pat == "" || cfg.UserID == "" {
			m.screen = screenDualPanel
			m.errMsg = "Apps list is empty. Set MX_PAT and UserID via 'mx config', then run 'mx sync'."
		} else {
			m.screen = screenLoading
			m.loadingMsg = "Syncing apps from Mendix Platform..."
		}
	} else {
		m.screen = screenDualPanel
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Commands

func syncCmd(ctx context.Context, cfg *config.Config, pat string) tea.Cmd {
	return func() tea.Msg {
		err := platform.Sync(ctx, cfg, pat, func(string) {})
		return syncDoneMsg{err: err}
	}
}

func getBranchesCmd(ctx context.Context, pat, appID string) tea.Cmd {
	return func() tea.Msg {
		branches, err := platform.GetBranches(ctx, pat, appID)
		return branchesLoadedMsg{branches: branches, err: err}
	}
}

func checkoutBranchCmd(ctx context.Context, cfg *config.Config, app config.App, branchName string, jobID int) tea.Cmd {
	return func() tea.Msg {
		safeBranch := strings.ReplaceAll(branchName, "/", "_")
		destDir := filepath.Join(cfg.ProjectDirectory, app.Name+"-"+safeBranch)
		var errBuf bytes.Buffer
		err := branch.Checkout(ctx, app, branchName, destDir, io.Discard, &errBuf)
		if err != nil {
			return branchOpDoneMsg{jobID: jobID, err: fmt.Errorf("%w\n%s", err, strings.TrimSpace(errBuf.String()))}
		}
		return branchOpDoneMsg{jobID: jobID, destDir: destDir}
	}
}

func createBranchCmd(ctx context.Context, cfg *config.Config, app config.App, branchName, baseBranch string, jobID int) tea.Cmd {
	return func() tea.Msg {
		var errBuf bytes.Buffer
		err := branch.Create(ctx, cfg, app, branchName, baseBranch, io.Discard, &errBuf)
		safeBranch := strings.ReplaceAll(branchName, "/", "_")
		destDir := filepath.Join(cfg.ProjectDirectory, app.Name+"-"+safeBranch)
		if err != nil {
			return branchOpDoneMsg{jobID: jobID, err: fmt.Errorf("%w\n%s", err, strings.TrimSpace(errBuf.String()))}
		}
		return branchOpDoneMsg{jobID: jobID, destDir: destDir}
	}
}

func clearJobAfterCmd(id int) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(4 * time.Second)
		return clearJobMsg{id: id}
	}
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spinner.Tick}
	if m.screen == screenLoading {
		cmds = append(cmds, syncCmd(m.ctx, m.cfg, m.pat))
	}
	return tea.Batch(cmds...)
}
