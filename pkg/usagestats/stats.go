package usagestats

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/prometheus/common/version"
)

var (
	httpClient    = http.Client{Timeout: 5 * time.Second}
	usageStatsURL = "https://stats.grafana.org/agent-usage-report"
)

// Report is the payload to be sent to stats.grafana.org
type Report struct {
	ClusterID string                 `json:"clusterID"`
	CreatedAt time.Time              `json:"createdAt"`
	Interval  time.Time              `json:"interval"`
	Version   string                 `json:"version"`
	Metrics   map[string]interface{} `json:"metrics"`
	Os        string                 `json:"os"`
	Arch      string                 `json:"arch"`
}

func sendReport(ctx context.Context, seed *ClusterSeed, interval time.Time, metrics map[string]interface{}) error {
	report := Report{
		ClusterID: seed.UID,
		CreatedAt: seed.CreatedAt,
		Version:   version.Print("agent"),
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Interval:  interval,
		Metrics:   metrics,
	}
	out, err := jsoniter.MarshalIndent(report, "", " ")
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, usageStatsURL, bytes.NewBuffer(out))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to send usage stats: %s  body: %s", resp.Status, string(data))
	}
	return nil
}