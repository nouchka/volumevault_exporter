package volumevault

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	volumesPath     = "/volumes"
	backupRunsPath  = "/backup-runs"
	restoreRunsPath = "/restore-runs"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type DockerVolume struct {
	ID                  int64   `json:"id"`
	Name                string  `json:"name"`
	BackupState         string  `json:"backup_state"`
	LastBackupSizeBytes *int64  `json:"last_backup_size_bytes"`
	LastBackupAt        *string `json:"last_backup_at"`
}

type BackupRun struct {
	ID              int64   `json:"id"`
	Status          string  `json:"status"`
	DurationSeconds *int64  `json:"duration_seconds"`
	BackupSizeBytes *int64  `json:"backup_size_bytes"`
	StartedAt       *string `json:"started_at"`
	FinishedAt      *string `json:"finished_at"`
}

type RestoreRun struct {
	ID         int64   `json:"id"`
	Status     string  `json:"status"`
	StartedAt  *string `json:"started_at"`
	FinishedAt *string `json:"finished_at"`
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) GetVolumes(ctx context.Context) ([]DockerVolume, error) {
	return getList[DockerVolume](ctx, c, volumesPath)
}

func (c *Client) GetBackupRuns(ctx context.Context) ([]BackupRun, error) {
	return getList[BackupRun](ctx, c, backupRunsPath)
}

func (c *Client) GetRestoreRuns(ctx context.Context) ([]RestoreRun, error) {
	return getList[RestoreRun](ctx, c, restoreRunsPath)
}

func getList[T any](ctx context.Context, c *Client, path string) ([]T, error) {
	body, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var direct []T
	if err := json.Unmarshal(body, &direct); err == nil {
		return direct, nil
	}

	var wrapped struct {
		Data []T `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && wrapped.Data != nil {
		return wrapped.Data, nil
	}

	return nil, fmt.Errorf("unable to decode response from %s", path)
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for %s: %w", path, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("%s returned HTTP %d: %s", path, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s response: %w", path, err)
	}

	return body, nil
}
