# TUI Feature Plan

Interactive terminal UI launched by `mx` with no arguments, built with [Charm](https://charm.sh/) (Bubble Tea + Lip Gloss + Bubbles).

---

## Resolved Decisions

- **`c` from app list** — uses the currently highlighted app directly; no extra selection step.
- **PAT not set** — TUI fails to open branch operations with an inline error; user is told to run `mx config` and set `MX_PAT`.
- **Empty apps list** — auto-trigger sync with a spinner on TUI open; show progress inline.
- **"Open branch"** — calls `project.Open()` (opens `.mpr` in Studio Pro). Confirmed.

---

## Dependencies to Add

```
github.com/charmbracelet/bubbletea   — TUI framework (Elm-style Model/Update/View)
github.com/charmbracelet/lipgloss    — Styling and layout
github.com/charmbracelet/bubbles     — Pre-built list, spinner, text-input components
```

---

## Config Change: Store AppID

The Mendix branches API requires the app's UUID (`AppID`), which is available during sync but currently discarded.

**`config/config.go`** — extend `App`:
```go
type App struct {
    Name          string
    AppID         string  // Mendix project UUID, needed for branches API
    RepositoryURL string
}
```

**`platform/sync.go`** — populate `AppID` when building the app list:
```go
apps = append(apps, config.App{
    Name:          proj.Name,
    AppID:         info.AppID,
    RepositoryURL: info.URL,
})
```

Existing configs without `AppID` will have an empty string; the TUI shows an error on branch operations for those apps and prompts the user to re-sync.

---

## New Platform Function: GetBranches

**`platform/platform.go`** — add:

```go
type Branch struct {
    Name string `json:"name"`
}

func GetBranches(ctx context.Context, pat, appID string) ([]Branch, error)
// GET https://repository.api.mendix.com/v1/repositories/<appID>/branches
```

Handles pagination the same way `GetUserProjects` does.

---

## New Package: `tui/`

```
tui/
├── tui.go       — Run() entry point, model definition, screen constants
├── update.go    — Update() and all message/key handling
├── view.go      — View() and per-screen render functions
└── styles.go    — Lip Gloss style definitions
```

---

## Top-Level Layout: Two Panels

The TUI has two top-level panels navigated horizontally:

```
[ Apps ]  <— default        [ Versions ]
```

- `l` / `→` from the apps panel moves to the versions panel
- `h` / `←` from the versions panel moves back to the apps panel
- The active panel is visually indicated (e.g. bold/underlined header)

---

## Screen States

```
appList          — Apps panel: list of apps from config.Apps (auto-syncs if empty)
versionList      — Versions panel: Studio Pro installations from VersionDirectory
branchList       — Local project dirs for the selected app (drilled into from appList)
branchAction     — "Create" / "Checkout" choice
remoteBranchList — Remote branches from Mendix API (base for create OR branch to checkout)
branchNameInput  — Text input for new branch name (create only)
loading          — Spinner shown during API calls and git operations
```

---

## Key Bindings (per screen)

| Key | appList | versionList | branchList | branchAction | remoteBranchList | branchNameInput |
|-----|---------|-------------|------------|--------------|------------------|-----------------|
| j / ↓ | down | down | down | down | down | — |
| k / ↑ | up | up | up | up | up | — |
| l / → | → versionList | — | — | — | — | — |
| h / ← | — | → appList | — | — | — | — |
| enter | → branchList | open version | open branch | confirm choice | confirm selection | confirm name |
| c | create/checkout | — | create/checkout | — | — | — |
| o | open config | — | — | — | — | — |
| esc | — | — | → appList | back | back | cancel |
| q | quit | quit | quit | quit | quit | quit |

---

## Footer Examples

```
appList:
  (j/k ↑↓) navigate | (l →) versions | (enter) view branches | (c) create/checkout | (o) open config | (q) quit

versionList:
  (j/k ↑↓) navigate | (h ←) apps | (enter) open version | (q) quit

branchList:
  (j/k ↑↓) navigate | (enter) open | (c) create/checkout | (esc) back | (q) quit

branchAction:
  (j/k ↑↓) navigate | (enter) select | (esc) back

remoteBranchList (checkout):
  (j/k ↑↓) navigate | (enter) checkout | (esc) back

remoteBranchList (create — pick base):
  (j/k ↑↓) navigate | (enter) use as base | (esc) back

branchNameInput:
  type branch name | (enter) create | (esc) cancel
```

---

## Flows

### Auto-sync on open

```
TUI opens → config.Apps is empty
  → loading: platform.Sync()
  → appList populated
```

### View local branches

```
appList (enter on app)
  → project.Search(ProjectDirectory, appName)
  → branchList (shows matching local dirs)
    (enter on branch) → project.Open() → quit TUI
```

### Open version

```
versionList (enter on version)
  → version.Open() → quit TUI
```

### Create branch

```
appList or branchList (c on highlighted app)
  → branchAction: [Create] [Checkout]
    (enter on Create)
    → loading: platform.GetBranches(appID)
    → remoteBranchList: pick base branch
      (enter on branch)
      → branchNameInput: type new branch name
        (enter) → loading: branch.Create(...) → open in Studio Pro → quit TUI
```

### Checkout branch

```
appList or branchList (c on highlighted app)
  → branchAction: [Create] [Checkout]
    (enter on Checkout)
    → loading: platform.GetBranches(appID)
    → remoteBranchList: pick branch to checkout
      (enter on branch) → loading: branch.Checkout(...) → open in Studio Pro → quit TUI
```

---

## Styling

- **Active panel header:** bold + underline
- **Focused list item:** grey background (`lipgloss.Color("240")`), default foreground (readable)
- **Footer:** dim/muted style, items separated by `|`
- **Error messages:** red foreground, shown inline above the footer
- **Spinner:** dots style (bubbles spinner) during loading

---

## Entry Point

**`main.go`** — root command `RunE` (currently shows help text):
```go
var rootCmd = &cobra.Command{
    Use: "mx",
    RunE: func(cmd *cobra.Command, args []string) error {
        return tui.Run(cfg)
    },
}
```

`tui.Run(cfg *config.Config) error` starts the Bubble Tea program.

---

## Implementation Order

1. Add `AppID` to `config.App`; update `sync.go`
2. Add `platform.GetBranches`
3. Add Charm dependencies (`go get`)
4. Create `tui/` package scaffolding (model, styles)
5. Implement `appList` + `versionList` panels with horizontal navigation
6. Implement auto-sync on empty apps (loading state)
7. Implement `branchList` screen (local project search)
8. Implement `branchAction` screen (Create/Checkout picker)
9. Implement `remoteBranchList` screen with loading state
10. Implement `branchNameInput` screen
11. Wire up `branch.Create` and `branch.Checkout` with progress feedback
12. Update `main.go` root command
13. Update `SPEC.md`
