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
