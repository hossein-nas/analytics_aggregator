package clarity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sync"
	"time"
)

type Collector struct {
	config  Config
	client  *http.Client
	metrics map[string]interface{}
	mu      sync.RWMutex
}

// Information struct for common fields
type Information struct {
	SessionsCount                   string  `json:"sessionsCount" bson:"sessions_count,omitempty"`
	SessionsWithMetricPercentage    float64 `json:"sessionsWithMetricPercentage" bson:"sessions_with_metric_percentage,omitempty"`
	SessionsWithoutMetricPercentage float64 `json:"sessionsWithoutMetricPercentage" bson:"sessions_without_metric_percentage,omitempty"`
	PagesViews                      string  `json:"pagesViews" bson:"pages_views,omitempty"`
	SubTotal                        string  `json:"subTotal" bson:"sub_total,omitempty"`
	AverageScrollDepth              float64 `json:"averageScrollDepth" bson:"average_scroll_depth,omitempty"`
	TotalSessionCount               string  `json:"totalSessionCount" bson:"total_session_count,omitempty"`
	TotalBotSessionCount            string  `json:"totalBotSessionCount" bson:"total_bot_session_count,omitempty"`
	DistinctUserCount               string  `json:"distinctUserCount" bson:"distinct_user_count,omitempty"`
	PagesPerSessionPercentage       float64 `json:"pagesPerSessionPercentage" bson:"pages_per_session_percentage,omitempty"`
	TotalTime                       string  `json:"totalTime" bson:"total_time,omitempty"`
	ActiveTime                      string  `json:"activeTime" bson:"active_time,omitempty"`
	Name                            string  `json:"name" bson:"name,omitempty"`
	URL                             string  `json:"url" bson:"url,omitempty"`
	VisitsCount                     string  `json:"visitsCount" bson:"visits_count,omitempty"`
}

// Metric struct for each metric
type Metric struct {
	MetricName  string        `json:"metricName" bson:"metric_name"`
	Information []Information `json:"information" bson:"information"`
}

// Response struct for the entire JSON response
type ClarityStats []Metric

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
	endpoint := "https://www.clarity.ms/export-data/api/v1/project-live-insights?numOfDays=1"

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
	fmt.Println("resp", resp)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Println("bodyBytes", string(bodyBytes))
	var stats ClarityStats
	if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&stats); err != nil {
		fmt.Println("err", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &stats, nil
}

func (c *Collector) updateMetrics(stats *ClarityStats) {
	c.mu.Lock()
	defer c.mu.Unlock()

	value, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println("stats", string(value))

	c.metrics = map[string]interface{}{
		"clarity_total_sessions":         getMetricValue(stats, "Traffic", "totalSessionCount"),
		"clarity_total_bot_sessions":     getMetricValue(stats, "Traffic", "totalBotSessionCount"),
		"clarity_distinct_users":         getMetricValue(stats, "Traffic", "distinctUserCount"),
		"clarity_pages_per_session":      getMetricValue(stats, "Traffic", "pagesPerSessionPercentage"),
		"clarity_scroll_depth_avg":       getMetricValue(stats, "ScrollDepth", "averageScrollDepth"),
		"clarity_total_engagement_time":  getMetricValue(stats, "EngagementTime", "totalTime"),
		"clarity_active_engagement_time": getMetricValue(stats, "EngagementTime", "activeTime"),
		"clarity_dead_clicks":            getMetricValue(stats, "DeadClickCount", "subTotal"),
		"clarity_rage_clicks":            getMetricValue(stats, "RageClickCount", "subTotal"),
		"clarity_excessive_scrolls":      getMetricValue(stats, "ExcessiveScroll", "subTotal"),
		"clarity_quickback_clicks":       getMetricValue(stats, "QuickbackClick", "subTotal"),
		"clarity_script_errors":          getMetricValue(stats, "ScriptErrorCount", "subTotal"),
		"clarity_error_clicks":           getMetricValue(stats, "ErrorClickCount", "subTotal"),
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

func getMetricValue(stats *ClarityStats, metricName string, fieldName string) interface{} {
	if stats == nil {
		return nil
	}
	for _, metric := range *stats {
		if metric.MetricName == metricName && len(metric.Information) > 0 {
			v := reflect.ValueOf(metric.Information[0])
			if f := v.FieldByName(fieldName); f.IsValid() {
				return f.Interface()
			}
		}
	}
	return nil
}
