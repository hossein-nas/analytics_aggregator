package appmetric

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Collector struct {
	config  Config
	client  *http.Client
	metrics map[string]interface{}
	mu      sync.RWMutex
}

type AppMetricStats struct {
	ResponseTime  float64 `json:"responseTime"`
	RequestCount  int     `json:"requestCount"`
	ErrorCount    int     `json:"errorCount"`
	CPUUsage      float64 `json:"cpuUsage"`
	MemoryUsage   float64 `json:"memoryUsage"`
	ActiveUsers   int     `json:"activeUsers"`
	DatabaseCalls int     `json:"databaseCalls"`
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
	if c.config.ApplicationID == "" {
		return fmt.Errorf("application ID is required")
	}
	if c.config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	return nil
}

func (c *Collector) Collect(ctx context.Context) error {
	baseURL := c.config.Host
	if baseURL == "" {
		baseURL = "https://api.appmetrics.io/v1"
	}

	endpoint := fmt.Sprintf("%s/applications/%s/metrics", baseURL, c.config.ApplicationID)
	stats, err := c.fetchStats(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch AppMetric stats: %w", err)
	}

	c.updateMetrics(stats)
	return nil
}

func (c *Collector) fetchStats(ctx context.Context, endpoint string) (*AppMetricStats, error) {
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

	var stats AppMetricStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &stats, nil
}

func (c *Collector) updateMetrics(stats *AppMetricStats) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics = map[string]interface{}{
		"app_response_time_ms":   stats.ResponseTime,
		"app_requests_total":     stats.RequestCount,
		"app_errors_total":       stats.ErrorCount,
		"app_cpu_usage_percent":  stats.CPUUsage,
		"app_memory_usage_bytes": stats.MemoryUsage,
		"app_active_users":       stats.ActiveUsers,
		"app_database_calls":     stats.DatabaseCalls,
		"app_error_rate":         float64(stats.ErrorCount) / float64(stats.RequestCount),
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
