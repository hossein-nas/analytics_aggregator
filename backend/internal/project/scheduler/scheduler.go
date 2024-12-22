package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hossein-nas/analytics_aggregator/internal/project/collector"
	model "github.com/hossein-nas/analytics_aggregator/internal/project/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProjectRepository interface {
	GetAllProjects(ctx context.Context) ([]model.Project, error)
}

type Scheduler struct {
	config      *Config
	projectRepo ProjectRepository
	collectors  map[string]collector.MetricsCollector
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

func NewScheduler(config *Config, projectRepo ProjectRepository) *Scheduler {
	return &Scheduler{
		config:      config,
		projectRepo: projectRepo,
		collectors:  make(map[string]collector.MetricsCollector),
		stopChan:    make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	ticker := time.NewTicker(s.config.CollectionInterval)
	defer ticker.Stop()

	// Initial collection
	if err := s.collectAllProjects(ctx); err != nil {
		log.Printf("Error in initial collection: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := s.collectAllProjects(ctx); err != nil {
				log.Printf("Error collecting metrics: %v", err)
			}
		case <-s.stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Scheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()
}

func (s *Scheduler) collectAllProjects(ctx context.Context) error {
	projects, err := s.projectRepo.GetAllProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to get projects: %w", err)
	}

	// Create a buffered channel to limit concurrent collections
	semaphore := make(chan struct{}, s.config.MaxWorkers)

	// Create a channel for collecting errors
	errChan := make(chan error, len(projects))

	// Start collection for each project
	for _, project := range projects {
		s.wg.Add(1)
		go func(p model.Project) {
			defer s.wg.Done()
			semaphore <- struct{}{}        // Acquire token
			defer func() { <-semaphore }() // Release token

			if err := s.collectProjectMetrics(ctx, p); err != nil {
				errChan <- fmt.Errorf("failed to collect metrics for project %s: %w", p.Name, err)
			}
		}(project)
	}

	// Wait for all collections to complete
	s.wg.Wait()
	close(errChan)

	// Collect all errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors during collection", len(errors))
	}

	return nil
}

func (s *Scheduler) collectProjectMetrics(ctx context.Context, project model.Project) error {
	for _, collectorType := range project.Collectors {
		collector, err := s.getCollectorForType(collectorType, project)
		if err != nil {
			return fmt.Errorf("failed to get collector: %w", err)
		}

		if err := collector.Collect(ctx); err != nil {
			return fmt.Errorf("failed to collect metrics for %s: %w", collectorType, err)
		}

		metrics, err := collector.GetMetrics()
		if err != nil {
			return fmt.Errorf("failed to get metrics for %s: %w", collectorType, err)
		}

		// Here you would store the metrics in your database
		if err := s.storeMetrics(ctx, project.ID, collectorType, metrics); err != nil {
			return fmt.Errorf("failed to store metrics for %s: %w", collectorType, err)
		}
	}

	return nil
}

func (s *Scheduler) getCollectorForType(collectorType string, project model.Project) (collector.MetricsCollector, error) {
	// Implementation depends on your collector factory/configuration
	// This is just an example
	switch collectorType {
	case "sentry":
		return createSentryCollector(project)
	case "clarity":
		return createClarityCollector(project)
	case "embrace":
		return createEmbraceCollector(project)
	case "appmetric":
		return createAppMetricCollector(project)
	default:
		return nil, fmt.Errorf("unknown collector type: %s", collectorType)
	}
}

func (s *Scheduler) storeMetrics(ctx context.Context, projectID primitive.ObjectID, collectorType string, metrics map[string]interface{}) error {
	// Implement metric storage logic here
	// This could be storing in MongoDB, sending to a time-series database, etc.
	return nil
}
