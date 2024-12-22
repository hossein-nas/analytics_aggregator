package clarity

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

type ClarityStats struct {
	Sessions    int     `json:"sessions"`
	PageViews   int     `json:"pageViews"`
	ScrollDepth float64 `json:"scrollDepth"`
	TimeOnSite  float64 `json:"timeOnSite"` // in seconds
	BounceRate  float64 `json:"bounceRate"`
}

func NewCollector(config Config) (*Collector, error) {
	collector := &Collector{
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

	if err := collector.Validate(); err != nil {
		return nil, err
	}

	return collector, nil
}

func (c *Collector) Validate() error {
	if c.config.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if c.config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	return nil
}

func (c *Collector) Collect(ctx context.Context) error {
	baseURL := c.config.Host
	if baseURL == "" {
		baseURL = "https://clarity.microsoft.com/api/v1"
	}

	endpoint := fmt.Sprintf("%s/projects/%s/metrics", baseURL, c.config.ProjectID)
	stats, err := c.fetchStats(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch Clarity stats: %w", err)
	}

	c.updateMetrics(stats)
	return nil
}

func (c *Collector) fetchStats(ctx context.Context, endpoint string) (*ClarityStats, error) {
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

	var stats ClarityStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &stats, nil
}

func (c *Collector) updateMetrics(stats *ClarityStats) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics = map[string]interface{}{
		"clarity_sessions_total":    stats.Sessions,
		"clarity_pageviews_total":   stats.PageViews,
		"clarity_scroll_depth_avg":  stats.ScrollDepth,
		"clarity_time_on_site_avg":  stats.TimeOnSite,
		"clarity_bounce_rate":       stats.BounceRate,
		"clarity_pages_per_session": float64(stats.PageViews) / float64(stats.Sessions),
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
