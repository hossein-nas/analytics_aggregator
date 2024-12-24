package sentry

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
	if c.config.OrganizationSlug == "" {
		return fmt.Errorf("organization slug is required")
	}
	if c.config.ProjectSlug == "" {
		return fmt.Errorf("project slug is required")
	}
	if c.config.AuthToken == "" {
		return fmt.Errorf("auth token is required")
	}
	return nil
}

type SentryStats struct {
	Start     time.Time   `json:"start"`
	End       time.Time   `json:"end"`
	Intervals []time.Time `json:"intervals,omitempty"`
	Groups    []Group     `json:"groups"`
}

// Group represents each group object in the array
type Group struct {
	By     By     `json:"by,omitempty"`
	Totals Totals `json:"totals,omitempty"`
	Series Series `json:"series,omitempty"`
}

// By represents the categorization of each group
type By struct {
	Category string `json:"category,omitempty"`
}

// Totals represents the totals object in each group
type Totals struct {
	SumQuantity int `json:"sum(quantity),omitempty"`
}

// Series represents the series object in each group
type Series struct {
	SumQuantity []int `json:"sum(quantity),omitempty"`
}

func (c *Collector) Collect(ctx context.Context) error {
	baseURL := c.config.Host
	if baseURL == "" {
		baseURL = "https://sentry.io"
	}

	// Get stats for last 24 hours
	now := time.Now()
	startOfDay := now.Truncate(24 * time.Hour)

	endpoint := fmt.Sprintf(
		"%s/api/0/organizations/%s/stats_v2/?start=%d&end=%d&field=%s&project_id=17&groupBy=%s&interval=1d",
		baseURL,
		c.config.OrganizationSlug,
		startOfDay.Unix(),
		now.Unix(),
		"sum(quantity)",
		"category",
	)

	stats, err := c.fetchStats(ctx, endpoint)
	fmt.Println(baseURL, endpoint, stats, err)
	if err != nil {
		return fmt.Errorf("failed to fetch stats: %w", err)
	}

	c.updateMetrics(stats)
	return nil
}

func (c *Collector) fetchStats(ctx context.Context, endpoint string) ([]SentryStats, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.AuthToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// // Read the response body
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Println("Error reading the response body:", err)
	// 	// return
	// }

	// // Convert the body to a string
	// bodyString := string(body)

	// // Print the human-readable response body
	// fmt.Println("Response body:\n", bodyString)

	var stats []SentryStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return stats, nil
}

func (c *Collector) updateMetrics(stats []SentryStats) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var totalEvents, totalErrors, totalCrashes int
	var recentEvents, recentErrors int

	// Calculate metrics for the last hour
	// oneHourAgo := time.Now().Add(-1 * time.Hour).Unix()

	// for _, stat := range stats {
	// 	totalEvents += stat.Count
	// 	totalErrors += stat.ErrorCount
	// 	totalCrashes += stat.CrashCount

	// 	if stat.Timestamp >= oneHourAgo {
	// 		recentEvents += stat.Count
	// 		recentErrors += stat.ErrorCount
	// 	}
	// }

	c.metrics = map[string]interface{}{
		"sentry_events_total":     totalEvents,
		"sentry_errors_total":     totalErrors,
		"sentry_crashes_total":    totalCrashes,
		"sentry_events_last_hour": recentEvents,
		"sentry_errors_last_hour": recentErrors,
		"sentry_error_rate":       float64(totalErrors) / float64(totalEvents),
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
