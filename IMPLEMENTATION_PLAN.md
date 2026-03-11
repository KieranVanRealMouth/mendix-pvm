# Implementation Plan: Mendix Platform API Integration (`mx sync`)

## Overview

This plan covers:
1. Extending the config with the user's email and synced app data (the PAT token is **not** stored in the config file — it is kept as a user-level environment variable).
2. Prompting the user for credentials on first run.
3. Creating a `platform` package for Mendix API calls.
4. Implementing the `mx sync` command.

---

## Step 1 — Extend the `Config` struct (`config/config.go`)

Add three new fields to the `Config` struct:

```go
type App struct {
    Name          string `json:"name"`
    RepositoryURL string `json:"repositoryUrl"`
}

type Config struct {
    VersionDirectory string `json:"versionDirectory"`
    ProjectDirectory string `json:"projectDirectory"`
    Email            string `json:"email"`
    Apps             []App  `json:"apps"`
}
```

- `Email` defaults to `""` (empty string) when skipped.
- The PAT token is **never written to the config file**. It is stored as the `MX_PAT` user-level environment variable (see Step 2).
- `Apps` is the persisted result of `mx sync`; defaults to `nil` / empty slice.
- Existing config files that omit these fields will deserialise with zero values — no migration needed.

---

## Step 2 — First-run credential prompting (`config/config.go`)

Modify `create()` so that after the directory defaults are resolved, the user is prompted interactively for `Email` and the PAT token.

### Prompt flow

1. Print an informational message, e.g.:
   ```
   To use Mendix Platform features (mx sync), you need a Personal Access Token (PAT).
   The PAT must include the following permissions:
     - mx:mxid3:user-identifiers:uuid:read
     - mx:app:metadata:read
     - mx:modelrepository:repo:read
     - mx:modelrepository:repo:write
     - mx:modelrepository:write

   The PAT will be stored as the MX_PAT environment variable — it will NOT be written to the config file.
   Press Enter to skip this step. These values can be set later by re-running the CLI or editing the config.
   ```
2. Read `Email` from stdin (`fmt.Scanln` / `bufio.Scanner`); save to config.
3. Read the PAT from stdin using `golang.org/x/term` to suppress echo.
4. If the PAT is non-empty, persist it as a **user-level environment variable** named `MX_PAT`:
   - **Windows:** run `setx MX_PAT <value>` via `os/exec`. Inform the user they must open a new terminal for it to take effect.
   - **Unix/Linux:** print instructions to add `export MX_PAT=<value>` to `~/.bashrc` / `~/.zshrc`.
5. If both inputs are blank (user presses Enter), leave `Email` as `""` and skip the env-var step — no error.

### Validation change

Update `validate()` so that an empty `Email` is **not** treated as an error. It is optional.

---

## Step 3 — Create `platform/` package

Create `platform/platform.go` (new package).

### 3.1 Shared HTTP helper

```go
func doRequest(ctx context.Context, pat, method, url string, body io.Reader) (*http.Response, error)
```

- Sets `Authorization: MxToken <pat>` header.
- Sets `Content-Type: application/json` when `body != nil`.
- Returns an error if the HTTP status is not 2xx.

The `pat` value is always sourced by the caller from `os.Getenv("MX_PAT")` — it is never read from or written to the config file.

### 3.2 Get user UUID — User Identifiers API

**Endpoint:** `POST https://user-management-api.home.mendix.com/v1/uuids`

**Request body:**
```json
{"emailAddresses": [{"emailAddress": "<email>"}]}
```

**Response (relevant part):**
```json
{
  "identifiers": [{"emailAddress": "...", "uuid": "..."}],
  "error": {"message": "...", "emailAddresses": [...]}
}
```

**Function signature:**
```go
func GetUserUUID(ctx context.Context, pat, email string) (string, error)
```

- Return an error if the response contains a non-empty `error.message` or if `identifiers` is empty.
- Return `identifiers[0].uuid`.

### 3.3 Get user projects — Projects API

**Endpoint:** `GET https://api.mendix.com/v1/users/<userId>/projects`

Supports pagination via `offset` and `limit` query parameters.

**Function signature:**
```go
func GetUserProjects(ctx context.Context, pat, userID string) ([]Project, error)
```

Where:
```go
type Project struct {
    ProjectID string `json:"projectId"`
    Name      string `json:"name"`
}
```

- Iterate pages until `page.offset + page.elements >= page.totalElements`.
- Collect all items across pages.

### 3.4 Get repository info — App Repository API

**Endpoint:** `GET https://repository.services.mendix.com/v1/repositories/<appId>/info`

**Response:**
```json
{"appId": "...", "type": "git|svn", "url": "..."}
```

**Function signature:**
```go
func GetRepositoryInfo(ctx context.Context, pat, appID string) (RepositoryInfo, error)
```

Where:
```go
type RepositoryInfo struct {
    AppID string `json:"appId"`
    Type  string `json:"type"`
    URL   string `json:"url"`
}
```

- Return an error on non-2xx responses.

---

## Step 4 — Add `Apps` persistence helpers to `config/config.go`

Add an exported method on `*Config` that the sync command can call to update and persist the app list:

```go
func (c *Config) SetApps(apps []App) error {
    c.Apps = apps
    return c.save()
}
```

---

## Step 5 — Implement the `mx sync` command (`main.go`)

Register a new `syncCmd` with cobra and add it to `rootCmd`.

### Command skeleton

```go
var syncCmd = &cobra.Command{
    Use:   "sync",
    Short: "Sync apps from the Mendix Platform",
    Long: `Fetch all accessible Mendix apps from the Mendix Platform and
store their names and Git repository URLs in the local config.

Requires a valid email address in the config and the MX_PAT environment
variable to be set. Run 'mx config' to set the email, or re-run the CLI
setup to configure both.`,
    RunE: func(cmd *cobra.Command, args []string) error { ... },
}
```

### `RunE` logic

```
1. pat := os.Getenv("MX_PAT")
   Check cfg.Email != "" && pat != ""
   → If either is empty, return a user-friendly error:
     "Mendix credentials are not configured. Ensure your email is set in
      the config ('mx config') and MX_PAT is set as an environment variable."

2. cmd.Println("Fetching user ID...")
   uuid, err := platform.GetUserUUID(ctx, pat, cfg.Email)

3. cmd.Println("Fetching projects...")
   projects, err := platform.GetUserProjects(ctx, pat, uuid)

4. For each project, call platform.GetRepositoryInfo(ctx, pat, project.ProjectID).
   - Skip on error (log a warning) or on type != "git".
   - Collect config.App{Name: project.Name, RepositoryURL: info.URL}.

5. cfg.SetApps(apps)

6. cmd.Printf("Synced %d Git app(s).\n", len(apps))
```

Consider running the per-project repository info calls concurrently (goroutines + `sync.WaitGroup`) to reduce wall-clock time for users with many projects, using a semaphore (buffered channel) to cap concurrency (e.g. 10).

---

## Step 6 — Update `go.mod` / dependencies

- `golang.org/x/term` — for reading the PAT token without echo during first-run setup (optional, can fall back to plain `fmt.Scan` if not desired).
- No new HTTP dependencies needed; use the standard `net/http` package.

Run:
```sh
go get golang.org/x/term
go mod tidy
```

---

## Step 7 — Update `rootCmd` help text (`main.go`)

Add `mx sync` to the `Long` description example block in `rootCmd`:

```
mx sync            Sync apps from the Mendix Platform
```

---

## File change summary

| File                   | Change                                                                                                                                                                                                                        |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `config/config.go`     | Add `App` struct; extend `Config` with `Email` and `Apps` (**no** `PATToken` field); update `create()` for credential prompting and `MX_PAT` env-var persistence; update `validate()` to allow empty `Email`; add `SetApps()` |
| `platform/platform.go` | New file — `doRequest` (reads PAT from caller, never from disk), `GetUserUUID`, `GetUserProjects`, `GetRepositoryInfo` and associated types                                                                                   |
| `main.go`              | Register `syncCmd` (reads `MX_PAT` via `os.Getenv`); update root help text                                                                                                                                                    |
| `go.mod` / `go.sum`    | Add `golang.org/x/term` for echo-suppressed PAT entry                                                                                                                                                                         |

---

## Sequence diagram

```
mx sync
  │
  ├─► platform.GetUserUUID(email, pat)
  │       POST /v1/uuids  →  uuid
  │
  ├─► platform.GetUserProjects(uuid, pat)
  │       GET /v1/users/<uuid>/projects  →  []Project  (paginated)
  │
  └─► for each project (concurrent):
          platform.GetRepositoryInfo(projectId, pat)
              GET /v1/repositories/<appId>/info
              skip if type != "git"
              →  {name, url}
  │
  └─► config.SetApps(apps)  →  persists to ~/.mendix-pvm.json
```
