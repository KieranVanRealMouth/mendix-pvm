package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func doRequest(ctx context.Context, pat, method, rawURL string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "MxToken "+pat)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %s: %s", resp.Status, rawURL)
	}
	return resp, nil
}

// --- GetUserProjects ---

type Project struct {
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
}

type pageInfo struct {
	Offset        int `json:"offset"`
	Elements      int `json:"elements"`
	TotalElements int `json:"totalElements"`
}

type projectsPage struct {
	Page  pageInfo  `json:"page"`
	Items []Project `json:"items"`
}

func GetUserProjects(ctx context.Context, pat, userID string) ([]Project, error) {
	const pageLimit = 20
	var all []Project
	for offset := 0; ; {
		u := fmt.Sprintf(
			"https://projects-api.home.mendix.com/v2/users/%s/projects?offset=%d&limit=%d",
			url.PathEscape(userID),
			offset,
			pageLimit,
		)
		resp, err := doRequest(ctx, pat, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		var page projectsPage
		if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		all = append(all, page.Items...)
		offset += len(page.Items)
		if offset >= page.Page.TotalElements {
			break
		}
	}
	return all, nil
}

// --- GetRepositoryInfo ---

type RepositoryInfo struct {
	AppID string `json:"appId"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

func GetRepositoryInfo(ctx context.Context, pat, appID string) (RepositoryInfo, error) {
	rawURL := fmt.Sprintf(
		"https://repository.api.mendix.com/v1/repositories/%s/info",
		url.PathEscape(appID),
	)
	resp, err := doRequest(ctx, pat, http.MethodGet, rawURL, nil)
	if err != nil {
		return RepositoryInfo{}, err
	}
	defer resp.Body.Close()

	var info RepositoryInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return RepositoryInfo{}, err
	}
	return info, nil
}
