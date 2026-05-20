---
tags:
  - technical specification
---

# Mendix PVM — Technical Specification

A CLI tool for managing Mendix Studio Pro versions and projects. Built in Go with Cobra for command parsing.

## Architecture Overview

The codebase follows a modular package-based architecture with clear separation of concerns:

```
main.go              — CLI command definitions and orchestration (Cobra)
├── config/          — Configuration management
├── version/         — Studio Pro version handling
├── project/         — Mendix project handling
├── platform/        — Mendix Platform API integration
├── branch/          — Git branch operations
├── convert/         — Project version conversion
├── search/          — Search and filtering logic
├── utils/           — Cross-platform utilities
└── ui/              — Output formatting
```

## Package Breakdown

### `config` — Configuration Management

**File:** `config/config.go`

Manages persistent configuration stored as JSON at `~/.mendix-pvm.json`.

**Data Structure:**

```go
type Config struct {
    VersionDirectory string  // Path to Studio Pro installations
    ProjectDirectory string  // Path to Mendix projects
    UserID           string  // Mendix OpenID for Platform API
    Apps             []App   // Synced Mendix apps from Platform
}

type App struct {
    Name          string  // App name
    RepositoryURL string  // Git repository URL
}
```

**Key Functions:**

- `Load()` — Loads config from home directory; creates default config on first run
- `create()` — Interactive setup wizard; guides users through initial configuration
- `validate()` — Validates directories exist and are readable
- `Open()` — Launches config file in platform-appropriate editor (cmd.exe, open, xdg-open)
- `persistPAT()` — Stores MX_PAT environment variable using platform-specific methods
- `SetApps()` — Updates and saves app list (used by sync)

**Platform-Specific Defaults:**

- Windows: versionDirectory = `%ProgramFiles%\Mendix`
- macOS/Linux: versionDirectory = empty (user must set)
- All platforms: projectDirectory = `~/Mendix`

**Design Decision:** Config stored in home directory rather than XDG_CONFIG_HOME for Windows compatibility. Future enhancement: full XDG support.

### `version` — Studio Pro Version Management

**File:** `version/version.go`

Handles discovery and launching of Studio Pro installations.

**Key Functions:**

- `Search(searchDirPath, args)` — Finds versions matching search terms; filters to only valid installations (must contain modeler subdirectory)
- `FindModelerSubdir(dir)` — Validates Studio Pro installation by checking for modeler/ directory
- `Open(versionPath)` — Launches Studio Pro with `--enable-extension-development` flag

**Implementation Details:**

- Version discovery is directory-based; searches top-level directories in versionDirectory
- Only returns paths containing a valid modeler/ subdirectory
- Launch method calls `studiopro.exe` from `<version>/modeler/studiopro.exe`

### `project` — Mendix Project Management

**File:** `project/project.go`

Handles Mendix project discovery and opening.

**Key Functions:**

- `Search(projectDirPath, args)` — Finds projects matching search terms; filters to only valid projects (must contain .mpr file at root)
- `FindMprAtRoot(projectPath)` — Validates project by checking for .mpr file in root directory
- `Open(projectPath)` — Launches project by opening .mpr file with system default application

**Implementation Details:**

- Project validation requires exactly one .mpr file at the directory root
- Opens .mpr files (not directly launching Studio Pro); OS handles launching appropriate application
- Uses `utils.OpenFile()` for cross-platform compatibility

### `search` — Search and Filtering

**File:** `search/search.go`

Provides normalized search across projects and versions.

**Key Functions:**

- `SearchDir(searchPath, query)` — Searches filesystem for directories matching all query tokens
- `SearchApps(apps, query)` — Searches app list (from config) by name
- `normalize(s)` — Converts strings to lowercase, removes non-alphanumeric characters
- `matchAllTokens(normName, tokens)` — Checks if all search tokens appear in normalized name

**Search Algorithm:**

1. Normalize query terms and directory/app names (lowercase, alphanumeric only)
2. Split query into space-separated tokens
3. Match if normalized directory/app name contains ALL tokens (order-independent, substring match)

**Example:** Query "ap 2" matches "MyApp v2.5" after normalization to "myappv25"

**Design Decision:** Substring matching prioritizes usability over strict matching; "app" finds "MyApp", "TheApplication", etc.

### `platform` — Mendix Platform API Integration

**Files:** `platform/platform.go`, `platform/sync.go`

Integrates with Mendix Platform APIs to discover and sync user applications.

**Key Functions:**

- `GetUserProjects(ctx, pat, userID)` — Fetches all projects user has access to; handles pagination
- `GetRepositoryInfo(ctx, pat, projectID)` — Fetches repository URL and type for a project
- `Sync(ctx, cfg, pat, printer)` — Orchestrates full sync: fetch projects → get repo info → filter Git repos → save config

**Sync Implementation:**

- Fetches projects first (paginated via offset/limit)
- Uses semaphore pattern to limit concurrent API requests to 10
- Filters to Git repos only (skips other repository types)
- Updates config.Apps with fetched repositories
- Non-blocking error handling: warns on individual failures but continues syncing other apps

**API Details:**

- Endpoint: Mendix platform APIs (exact URLs in code)
- Authentication: `MxToken <PAT>` header
- Response format: JSON with pagination metadata

**Design Decision:** Semaphore-based concurrency limits API load while fetching repository info for each project; Pat is required and not persisted (passed via environment variable MX_PAT).

### `branch` — Git Branch Management

**File:** `branch/branch.go`

Manages cloning and creating Git branches from Mendix Platform repositories.

**Key Functions:**

- `Checkout(ctx, app, branchName, destDir, stdout, stderr)` — Clones single branch from repository
- `Create(ctx, cfg, app, branchName, baseRef, stdout, stderr)` — Creates branch on remote, then clones it
- `setProjectId(destDir)` — Updates `.git/config` with Mendix project ID for private version control

**Checkout Flow:**

1. Clone single branch with `git clone --branch <branch> --single-branch <repo> <destDir>`
2. Extract and set Mendix project ID in git config (Mendix-hosted repos only)

**Create Flow:**

1. Check if branch exists on remote; create if not (via git push)
2. Clone the branch using checkout flow
3. Optionally open in Studio Pro if --open flag set

**Project ID Extraction:**

- Parses `.git/config` to extract remote URL
- Only applies to Mendix-hosted repos (https://git.api.mendix.com)
- Extracts last path segment as project ID
- Required for Studio Pro private version control integration

**Design Decision:** Separate checkout and create flows allow reuse; creates or reuses branches transparently to user; project ID extraction enables seamless Studio Pro Git integration.

### `convert` — Project Version Conversion

**File:** `convert/convert.go`

Delegates project conversion to Studio Pro's mx.exe tool.

**Key Function:**

- `Convert(versionPath, projectPath)` — Executes mx.exe with conversion arguments; returns exit code and error

**Exit Codes:**

- 0: Success
- 1: Internal error during conversion
- 2: Error with options
- 3: Conversion failed
- -1: Process start/execution error

**Implementation:** Wrapper around `mx.exe convert --in-place <projectPath>` from the specified Studio Pro version; streams stdout/stderr directly to user.

### `utils` — Cross-Platform Utilities

**File:** `utils/file.go`

Provides platform-specific file/application launching.

**Key Function:**

- `OpenFile(path)` — Launches file with system default application
  - Windows: `cmd /c start "" <path>`
  - macOS: `open <path>`
  - Linux: `xdg-open <path>`

Uses `runtime.GOOS` for platform detection; non-blocking (cmd.Start, not cmd.Run).

### `ui` — Output Formatting

**File:** `ui/list.go`

Formats search results for display.

**Key Function:**

- `List(items)` — Converts file paths to simple list format (directory names only, not full paths)

## Command Orchestration

**File:** `main.go` (689 lines)

Uses Cobra framework for command parsing and execution.

**Architecture:**

- Root command defines global help text
- Each major feature (list, open, path, convert, config, sync, branch) is a separate command
- Branch has nested subcommands (checkout, create)
- Flags are declared per-command using Cobra's flag API

**Flag Handling:**

- `--project/-p` and `--version/-v` are mutually exclusive options across commands
- `--all/-a` limits result count on multi-match operations
- Flags parsed by Cobra; values stored in command-scoped variables

**Key Design Patterns:**

- `searchTargets()` function encapsulates search logic for commands with project/version filtering
- `resolveTerminalPath()` determines if path is project or version
- `OpenProjectOrVersion()` dispatches to appropriate opener based on path type

## Configuration Flow

1. **Load:** `config.Load()` → read JSON from `~/.mendix-pvm.json`
2. **Create (first run):** Interactive wizard → prompt for directories → platform defaults → save JSON
3. **Validate:** Check directories exist and are readable
4. **Update (sync):** `Sync()` fetches from Platform → updates config.Apps → calls `config.SetApps()`
5. **Edit:** User runs `mx config` → opens file in editor → validation on next load

## Data Flow Examples

### List Command

```
user input: mx list 10.6 -v
  ↓
parse args (--version flag, "10.6" token)
  ↓
search.SearchDir(cfg.VersionDirectory, ["10.6"])
  ↓
normalize "10.6" → "106"
  ↓
list directories, filter by normalized name contains "106"
  ↓
version.Search() → filter to only valid installations (has modeler/)
  ↓
ui.List() → format as bullet list
  ↓
display to user
```

### Sync Command

```
user input: mx sync
  ↓
platform.Sync(ctx, cfg, pat, printer)
  ↓
GetUserProjects(pat, cfg.UserID)
  ↓
for each project: GetRepositoryInfo(pat, projectID)
  ↓
filter type==git, build config.App list
  ↓
cfg.SetApps(apps) → save to JSON
  ↓
config persisted; available for branch commands
```

### Branch Create Command

```
user input: mx branch create -r "Approval" -b feat/new --base main
  ↓
search.SearchApps(cfg.Apps, "Approval")
  ↓
if not found: auto-sync and retry
  ↓
branch.Create(ctx, cfg, app, "feat/new", "main")
  ↓
git push origin main:feat/new (creates branch)
  ↓
git clone --branch feat/new --single-branch <repo> <destDir>
  ↓
setProjectId(destDir) → update .git/config
  ↓
branch cloned and ready to open
```

## Error Handling Strategy

- **Validation errors:** Fail fast with clear messages (missing directories, invalid config)
- **Search not found:** Return empty results; commands handle gracefully (display "No matches found")
- **Sync errors:** Non-blocking per-app; warn on failures but continue syncing remaining apps
- **Platform API errors:** Return HTTP status code in error message
- **Git operations:** Stream stderr to user for debugging

## Future Enhancements

- **XDG_CONFIG_HOME Support:** Full XDG Base Directory specification for Linux/macOS
- **Multiple Project Directories:** Config extension to support array of project paths
- **Result Caching:** Cache search results with invalidation on filesystem changes
- **TUI Interface:** Interactive terminal UI on base `mx` command (no arguments)
- **Concurrent Conversions:** Parallel conversion support for multiple projects
