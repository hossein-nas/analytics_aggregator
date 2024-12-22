package embrace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Config struct {
	AppID  string
	APIKey string
	Host   string // Optional, defaults to Embrace API endpoint
}

type Collector struct {
	config  Config
	client  *http.Client
	metrics map[string]interface{}
	mu      sync.RWMutex
}

type EmbraceStats struct {
	Crashes      int     `json:"crashes"`
	ANRs         int     `json:"anrs"` // Application Not Responding
	NetworkCalls int     `json:"networkCalls"`
	ErrorRate    float64 `json:"errorRate"`
	SessionCount int     `json:"sessionCount"`
	UserCount    int     `json:"userCount"`
}

func NewCollector(config Config) *Collector {
	return &Collector{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: true,
			},
		},
		metrics: make(map[string]interface{}),
	}
}

func (c *Collector) Validate() error {
	if c.config.AppID == "" {
		return fmt.Errorf("app ID is required")
	}
	if c.config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	return nil
}

func (c *Collector) Collect(ctx context.Context) error {
	baseURL := c.config.Host
	if baseURL == "" {
		baseURL = "https://api.embrace.io/v1"
	}

	endpoint := fmt.Sprintf("%s/apps/%s/metrics", baseURL, c.config.AppID)
	stats, err := c.fetchStats(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch Embrace stats: %w", err)
	}

	c.updateMetrics(stats)
	return nil
}

func (c *Collector) fetchStats(ctx context.Context, endpoint string) (*EmbraceStats, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stats EmbraceStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &stats, nil
}

func (c *Collector) updateMetrics(stats *EmbraceStats) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics = map[string]interface{}{
		"embrace_crashes_total":    stats.Crashes,
		"embrace_anrs_total":       stats.ANRs,
		"embrace_network_calls":    stats.NetworkCalls,
		"embrace_error_rate":       stats.ErrorRate,
		"embrace_sessions_total":   stats.SessionCount,
		"embrace_users_total":      stats.UserCount,
		"embrace_crashes_per_user": float64(stats.Crashes) / float64(stats.UserCount),
	}
}

func (c *Collector) GetMetrics() (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics := make(map[string]interface{}, len(c.metrics))
	for k, v := range c.metrics {
		metrics[k] = v
	}
	return metrics, nil
}
