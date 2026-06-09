package exporter

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/nouchka/volumevault_exporter/internal/volumevault"

	"github.com/prometheus/client_golang/prometheus"
)

type API interface {
	GetVolumes(ctx context.Context) ([]volumevault.DockerVolume, error)
	GetBackupRuns(ctx context.Context) ([]volumevault.BackupRun, error)
	GetRestoreRuns(ctx context.Context) ([]volumevault.RestoreRun, error)
}

type Collector struct {
	api API

	up                  *prometheus.Desc
	scrapeDuration      *prometheus.Desc
	volumesTotal        *prometheus.Desc
	volumeBackupState   *prometheus.Desc
	volumeLastBackupTs  *prometheus.Desc
	volumeLastBackupSz  *prometheus.Desc
	backupRunsTotal     *prometheus.Desc
	restoreRunsTotal    *prometheus.Desc
	lastRunDurationSecs *prometheus.Desc
}

func NewCollector(api API) *Collector {
	return &Collector{
		api: api,
		up: prometheus.NewDesc(
			"volumevault_up",
			"Whether scraping VolumeVault API succeeded.",
			nil, nil,
		),
		scrapeDuration: prometheus.NewDesc(
			"volumevault_scrape_duration_seconds",
			"Duration of a VolumeVault scrape.",
			nil, nil,
		),
		volumesTotal: prometheus.NewDesc(
			"volumevault_volumes_total",
			"Number of Docker volumes known by VolumeVault.",
			nil, nil,
		),
		volumeBackupState: prometheus.NewDesc(
			"volumevault_volume_backup_state",
			"Backup state of each volume (1 for current state).",
			[]string{"volume", "state"}, nil,
		),
		volumeLastBackupTs: prometheus.NewDesc(
			"volumevault_volume_last_backup_timestamp_seconds",
			"Last backup timestamp of each volume in Unix seconds.",
			[]string{"volume"}, nil,
		),
		volumeLastBackupSz: prometheus.NewDesc(
			"volumevault_volume_last_backup_size_bytes",
			"Last backup size in bytes for each volume.",
			[]string{"volume"}, nil,
		),
		backupRunsTotal: prometheus.NewDesc(
			"volumevault_backup_runs_total",
			"Number of backup runs by status.",
			[]string{"status"}, nil,
		),
		restoreRunsTotal: prometheus.NewDesc(
			"volumevault_restore_runs_total",
			"Number of restore runs by status.",
			[]string{"status"}, nil,
		),
		lastRunDurationSecs: prometheus.NewDesc(
			"volumevault_backup_run_duration_seconds",
			"Duration in seconds for backup runs with known duration.",
			[]string{"run_id", "status"}, nil,
		),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.scrapeDuration
	ch <- c.volumesTotal
	ch <- c.volumeBackupState
	ch <- c.volumeLastBackupTs
	ch <- c.volumeLastBackupSz
	ch <- c.backupRunsTotal
	ch <- c.restoreRunsTotal
	ch <- c.lastRunDurationSecs
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var (
		volumes     []volumevault.DockerVolume
		backupRuns  []volumevault.BackupRun
		restoreRuns []volumevault.RestoreRun
		volErr      error
		backupErr   error
		restoreErr  error
	)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		volumes, volErr = c.api.GetVolumes(ctx)
	}()

	go func() {
		defer wg.Done()
		backupRuns, backupErr = c.api.GetBackupRuns(ctx)
	}()

	go func() {
		defer wg.Done()
		restoreRuns, restoreErr = c.api.GetRestoreRuns(ctx)
	}()

	wg.Wait()

	success := volErr == nil && backupErr == nil && restoreErr == nil
	if success {
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)
		c.collectVolumes(ch, volumes)
		c.collectBackupRuns(ch, backupRuns)
		c.collectRestoreRuns(ch, restoreRuns)
	} else {
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
	}

	ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, time.Since(start).Seconds())
}

func (c *Collector) collectVolumes(ch chan<- prometheus.Metric, volumes []volumevault.DockerVolume) {
	ch <- prometheus.MustNewConstMetric(c.volumesTotal, prometheus.GaugeValue, float64(len(volumes)))
	for _, v := range volumes {
		ch <- prometheus.MustNewConstMetric(c.volumeBackupState, prometheus.GaugeValue, 1, v.Name, v.BackupState)
		if v.LastBackupSizeBytes != nil {
			ch <- prometheus.MustNewConstMetric(c.volumeLastBackupSz, prometheus.GaugeValue, float64(*v.LastBackupSizeBytes), v.Name)
		}
		if v.LastBackupAt != nil {
			if ts, err := time.Parse(time.RFC3339, *v.LastBackupAt); err == nil {
				ch <- prometheus.MustNewConstMetric(c.volumeLastBackupTs, prometheus.GaugeValue, float64(ts.Unix()), v.Name)
			}
		}
	}
}

func (c *Collector) collectBackupRuns(ch chan<- prometheus.Metric, runs []volumevault.BackupRun) {
	counts := make(map[string]int)
	for _, run := range runs {
		counts[run.Status]++
		if run.DurationSeconds != nil {
			ch <- prometheus.MustNewConstMetric(c.lastRunDurationSecs, prometheus.GaugeValue, float64(*run.DurationSeconds), intToString(run.ID), run.Status)
		}
	}
	for status, count := range counts {
		ch <- prometheus.MustNewConstMetric(c.backupRunsTotal, prometheus.GaugeValue, float64(count), status)
	}
}

func (c *Collector) collectRestoreRuns(ch chan<- prometheus.Metric, runs []volumevault.RestoreRun) {
	counts := make(map[string]int)
	for _, run := range runs {
		counts[run.Status]++
	}
	for status, count := range counts {
		ch <- prometheus.MustNewConstMetric(c.restoreRunsTotal, prometheus.GaugeValue, float64(count), status)
	}
}

func intToString(v int64) string {
	return strconv.FormatInt(v, 10)
}
