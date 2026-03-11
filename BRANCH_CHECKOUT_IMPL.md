# Implementation Plan: `mx branch checkout`

## Overview

Add a `mx branch checkout` command that clones a single branch from an app's remote Git repository into the user's projects directory, mirroring how Mendix Studio Pro handles branch checkouts.

**Command signature:**
```
mx branch checkout --repository <query> --branch <branch-name>
mx branch checkout -r <query> -b <branch-name>
```

---

## Steps

### Step 1 — Export app-search helpers from the `search` package

`search/search.go` currently has unexported `normalize` and `matchAllTokens` helpers. App searching needs the same normalization logic but operates over `[]config.App` rather than filesystem directories.

Add a new exported function to `search/search.go`:

```go
// SearchApps returns every app whose name matches all tokens derived from query.
func SearchApps(apps []config.App, query string) []config.App {
    tokens := []string{}
    for _, q := range strings.Fields(query) {
        if n := normalize(q); n != "" {
            tokens = append(tokens, n)
        }
    }

    var matches []config.App
    for _, app := range apps {
        if matchAllTokens(normalize(app.Name), tokens) {
            matches = append(matches, app)
        }
    }
    return matches
}
```

- Add `"mendix-pvm/config"` to the import block in `search/search.go`.
- No import cycle is introduced: `config` has no dependency on `search`.

---

### Step 2 — Extract the sync logic into `platform/sync.go`

`syncCmd.RunE` in `main.go` contains the full sync logic inline. The checkout command needs to trigger a sync when an app is not found in the current config. Extract the body into a new file `platform/sync.go` so it lives alongside the other platform API code and can be called from multiple places.

Create `platform/sync.go`:

```go
package platform

import (
    "context"
    "fmt"
    "mendix-pvm/config"
    "sync"
)

// Sync fetches all Git apps from the Mendix Platform and updates cfg.
// printer is called for informational messages (e.g. cmd.Println).
func Sync(ctx context.Context, cfg *config.Config, pat string, printer func(string)) error {
    printer("Fetching projects...")
    projects, err := GetUserProjects(ctx, pat, cfg.UserID)
    if err != nil {
        return fmt.Errorf("failed to fetch projects: %w", err)
    }

    var (
        mu        sync.Mutex
        apps      []config.App
        wg        sync.WaitGroup
        semaphore = make(chan struct{}, 10)
    )

    for _, p := range projects {
        wg.Add(1)
        go func(proj Project) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            info, err := GetRepositoryInfo(ctx, pat, proj.ProjectID)
            if err != nil {
                printer(fmt.Sprintf("Warning: could not fetch repository info for %q: %v", proj.Name, err))
                return
            }
            if info.Type != "git" {
                return
            }

            mu.Lock()
            apps = append(apps, config.App{Name: proj.Name, RepositoryURL: info.URL})
            mu.Unlock()
        }(p)
    }
    wg.Wait()

    if err := cfg.SetApps(apps); err != nil {
        return fmt.Errorf("failed to save apps: %w", err)
    }
    printer(fmt.Sprintf("Synced %d Git app(s).", len(apps)))
    return nil
}
```

Update `syncCmd.RunE` in `main.go` to delegate to `platform.Sync`:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    pat := os.Getenv("MX_PAT")
    if cfg.UserID == "" || pat == "" {
        return fmt.Errorf("Mendix credentials are not configured. ...")
    }
    return platform.Sync(cmd.Context(), cfg, pat, func(s string) { cmd.Println(s) })
},
```

Remove any now-unused `sync` and `context` imports from `main.go` if the only remaining callers were the inline sync logic (they will be re-added via `platform.Sync`).

---

### Step 3 — Add the `branch` parent command

`mx branch` acts as a namespace group with no own `RunE`. Add to `main.go` inside `main()`:

```go
var branchCmd = &cobra.Command{
    Use:   "branch",
    Short: "Manage Mendix app branches",
    Long: `Commands for working with Mendix application branches.

Commands:
    checkout    Clone a branch from a Mendix app repository

Examples:
    mx branch checkout -r "Approval Tool" -b feat/my-feature
`,
}
```

---

### Step 4 — Add the `checkout` subcommand

Declare flag variables near the other flag variable declarations in `main()`:

```go
var (
    branchRepository string
    branchName       string
)
```

Define `checkoutCmd` inside `main()`:

```go
var checkoutCmd = &cobra.Command{
    Use:   "checkout",
    Short: "Clone a branch from a Mendix app's remote repository",
    Long: `Clone a single branch from a Mendix app's remote Git repository
into a new directory inside your configured projects directory.

The repository flag accepts a search query; the app name must contain the
provided string. If no matching app is found in the local config, a sync is
performed automatically before retrying.

Branch directories are created as:
    <App Name>-<branch_name>   (forward slashes in branch names become underscores)

Required:
    --repository, -r   Search query to identify the app
    --branch, -b       Branch name to clone

Examples:
    mx branch checkout -r "Approval Tool" -b feat/my-feature
    mx branch checkout --repository "Order" --branch main
`,
    RunE: func(cmd *cobra.Command, args []string) error {
        pat := os.Getenv("MX_PAT")

        // 1. Search current config
        matches := search.SearchApps(cfg.Apps, branchRepository)

        // 2. If no match, sync and retry
        if len(matches) == 0 {
            if cfg.UserID == "" || pat == "" {
                return fmt.Errorf("repository %q not found in config and Mendix credentials are not set; cannot sync", branchRepository)
            }
            cmd.Printf("Repository %q not found in config. Syncing...\n", branchRepository)
            if err := platform.Sync(cmd.Context(), cfg, pat, func(s string) { cmd.Println(s) }); err != nil {
                return fmt.Errorf("sync failed: %w", err)
            }
            matches = search.SearchApps(cfg.Apps, branchRepository)
        }

        // 3. Exit conditions
        if len(matches) == 0 {
            return fmt.Errorf("no repository found matching %q", branchRepository)
        }
        if len(matches) > 1 {
            var names []string
            for _, m := range matches {
                names = append(names, m.Name)
            }
            return fmt.Errorf("multiple repositories match %q: %s\nRefine your --repository query", branchRepository, strings.Join(names, ", "))
        }

        app := matches[0]

        // 4. Build destination directory path
        safeBranch := strings.ReplaceAll(branchName, "/", "_")
        dirName := app.Name + "-" + safeBranch
        destDir := filepath.Join(cfg.ProjectDirectory, dirName)

        // 5. Clone
        cmd.Printf("Cloning branch %q of %q into:\n  %s\n", branchName, app.Name, destDir)
        gitCmd := exec.CommandContext(
            cmd.Context(),
            "git", "clone",
            "--branch", branchName,
            "--single-branch",
            app.RepositoryURL,
            destDir,
        )
        gitCmd.Stdout = cmd.OutOrStdout()
        gitCmd.Stderr = cmd.ErrOrStderr()
        if err := gitCmd.Run(); err != nil {
            return fmt.Errorf("git clone failed: %w", err)
        }

        cmd.Printf("Done. Branch available at: %s\n", destDir)
        return nil
    },
}
```

---

### Step 5 — Bind flags and register commands

Add at the bottom of the flag-binding block in `main()` (before `rootCmd.Execute()`):

```go
checkoutCmd.Flags().StringVarP(&branchRepository, "repository", "r", "", "Search query to identify the app repository")
checkoutCmd.Flags().StringVarP(&branchName, "branch", "b", "", "Branch name to clone")
_ = checkoutCmd.MarkFlagRequired("repository")
_ = checkoutCmd.MarkFlagRequired("branch")

branchCmd.AddCommand(checkoutCmd)
rootCmd.AddCommand(branchCmd)
```

---

### Step 6 — Add required imports to `main.go`

Ensure the import block contains:

```go
import (
    // existing imports ...
    "os/exec"     // for exec.CommandContext
    "path/filepath"

    "mendix-pvm/search"
    // other existing project imports ...
)
```

`filepath` may already be present indirectly via other packages; verify before adding.

---

### Step 7 — Verify & test

```
go build ./...
```

Smoke-test the happy path (assuming a configured environment):

```
mx branch checkout -r "MyApp" -b main
```

Additional cases to verify manually:

| Scenario                             | Expected behaviour                                    |
| ------------------------------------ | ----------------------------------------------------- |
| App found in config, valid branch    | `git clone` runs, dir created in `ProjectDirectory`   |
| App not in config, valid credentials | Sync runs, then `git clone`                           |
| App not in config, no credentials    | Error: cannot sync                                    |
| Query matches multiple apps          | Error: listing matched names, asking for refinement   |
| Query matches no app after sync      | Error: exit                                           |
| Branch name contains `/`             | Dir uses `_` separator (e.g. `MyApp-feat_my-feature`) |

---

## File Change Summary

| File               | Change                                                                                                                                                                         |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `platform/sync.go` | **New file.** Exported `Sync(ctx, cfg, pat, printer)` function containing the extracted sync logic                                                                             |
| `search/search.go` | Add `SearchApps(apps []config.App, query string) []config.App`; add `config` import                                                                                            |
| `main.go`          | Replace inline sync body in `syncCmd.RunE` with `platform.Sync(...)`; add `branchCmd`, `checkoutCmd`; bind flags; register commands; add `os/exec` and `path/filepath` imports |
