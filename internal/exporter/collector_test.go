package exporter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nouchka/volumevault_exporter/internal/volumevault"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCollectorExposesExpectedMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/volumes":
			_, _ = w.Write([]byte(`{"data":[{"id":1,"name":"db","backup_state":"backed_up","last_backup_size_bytes":1024,"last_backup_at":"2026-06-09T15:00:00Z"}]}`))
		case "/backup-runs":
			_, _ = w.Write([]byte(`{"data":[{"id":10,"status":"success","duration_seconds":12},{"id":11,"status":"failed","duration_seconds":7}]}`))
		case "/restore-runs":
			_, _ = w.Write([]byte(`{"data":[{"id":21,"status":"queued"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	client := volumevault.NewClient(ts.URL, "token", 2*time.Second)
	collector := NewCollector(client)
	registry := prometheus.NewRegistry()
	if err := registry.Register(collector); err != nil {
		t.Fatalf("register collector: %v", err)
	}

	want := strings.NewReader(`# HELP volumevault_backup_runs_total Number of backup runs by status.
# TYPE volumevault_backup_runs_total gauge
volumevault_backup_runs_total{status="failed"} 1
volumevault_backup_runs_total{status="success"} 1
# HELP volumevault_restore_runs_total Number of restore runs by status.
# TYPE volumevault_restore_runs_total gauge
volumevault_restore_runs_total{status="queued"} 1
# HELP volumevault_up Whether scraping VolumeVault API succeeded.
# TYPE volumevault_up gauge
volumevault_up 1
# HELP volumevault_volume_backup_state Backup state of each volume (1 for current state).
# TYPE volumevault_volume_backup_state gauge
volumevault_volume_backup_state{state="backed_up",volume="db"} 1
# HELP volumevault_volume_last_backup_size_bytes Last backup size in bytes for each volume.
# TYPE volumevault_volume_last_backup_size_bytes gauge
volumevault_volume_last_backup_size_bytes{volume="db"} 1024
# HELP volumevault_volumes_total Number of Docker volumes known by VolumeVault.
# TYPE volumevault_volumes_total gauge
volumevault_volumes_total 1
`)

	if err := testutil.GatherAndCompare(registry, want,
		"volumevault_backup_runs_total",
		"volumevault_restore_runs_total",
		"volumevault_up",
		"volumevault_volume_backup_state",
		"volumevault_volume_last_backup_size_bytes",
		"volumevault_volumes_total",
	); err != nil {
		t.Fatalf("metrics mismatch: %v", err)
	}
}

func TestCollectorDownWhenEndpointFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/backup-runs" {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer ts.Close()

	client := volumevault.NewClient(ts.URL, "token", 2*time.Second)
	collector := NewCollector(client)
	reg := prometheus.NewRegistry()
	if err := reg.Register(collector); err != nil {
		t.Fatalf("register collector: %v", err)
	}

	want := strings.NewReader(`# HELP volumevault_up Whether scraping VolumeVault API succeeded.
# TYPE volumevault_up gauge
volumevault_up 0
`)
	if err := testutil.GatherAndCompare(reg, want, "volumevault_up"); err != nil {
		t.Fatalf("metrics mismatch: %v", err)
	}
}

var _ API = (*fakeAPI)(nil)

type fakeAPI struct{}

func (f *fakeAPI) GetVolumes(ctx context.Context) ([]volumevault.DockerVolume, error) {
	return nil, nil
}
func (f *fakeAPI) GetBackupRuns(ctx context.Context) ([]volumevault.BackupRun, error) {
	return nil, nil
}
func (f *fakeAPI) GetRestoreRuns(ctx context.Context) ([]volumevault.RestoreRun, error) {
	return nil, nil
}
