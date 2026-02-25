package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	Enabled                   bool   `env:"ENABLE_SCHEDULER" envDefault:"true"`
	EmployeeSyncSchedule      string `env:"EMPLOYEE_SYNC_SCHEDULE" envDefault:"0 0 * * 0"` // Weekly on Sunday
	GitActivitiesSyncSchedule string `env:"GIT_SYNC_SCHEDULE" envDefault:"0 */6 * * *"`    // Every 6 hours
	AggregationSchedule       string `env:"AGGREGATION_SCHEDULE" envDefault:"0 2 * * *"`   // Daily at 2 AM
}

// Scheduler handles periodic job execution
type Scheduler struct {
	cron               *cron.Cron
	config             SchedulerConfig
	aggregationService *AggregationService
	employeeService    *EmployeeService // Will be implemented
	gitSyncService     *GitSyncService  // Will be implemented
}

// NewScheduler creates a new scheduler
func NewScheduler(
	config SchedulerConfig,
	aggregationService *AggregationService,
) *Scheduler {
	return &Scheduler{
		cron:               cron.New(cron.WithSeconds()),
		config:             config,
		aggregationService: aggregationService,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	if !s.config.Enabled {
		slog.Info("Scheduler is disabled")
		return nil
	}

	slog.Info("Starting scheduler with configured jobs")

	// Register aggregation job
	if err := s.registerAggregationJob(); err != nil {
		return fmt.Errorf("failed to register aggregation job: %w", err)
	}

	// Register employee sync job (if configured)
	if s.config.EmployeeSyncSchedule != "" {
		if err := s.registerEmployeeSyncJob(); err != nil {
			slog.Warn("Failed to register employee sync job", "error", err)
		}
	}

	// Register git activities sync job (if configured)
	if s.config.GitActivitiesSyncSchedule != "" {
		if err := s.registerGitActivitiesSyncJob(); err != nil {
			slog.Warn("Failed to register git activities sync job", "error", err)
		}
	}

	// Start the cron scheduler
	s.cron.Start()

	slog.Info("Scheduler started successfully")
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	slog.Info("Stopping scheduler")
	ctx := s.cron.Stop()
	<-ctx.Done()
	slog.Info("Scheduler stopped")
}

// registerAggregationJob registers the metrics aggregation job
func (s *Scheduler) registerAggregationJob() error {
	jobFunc := func() {
		slog.Info("Starting scheduled aggregation job")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		startTime := time.Now()

		// Aggregate for current month
		if err := s.aggregationService.AggregateForCurrentMonth(ctx); err != nil {
			slog.Error("Aggregation job failed", "error", err, "duration", time.Since(startTime))
			return
		}

		// Also aggregate for last 30 days for rolling metrics
		if err := s.aggregationService.AggregateForLastNDays(ctx, 30); err != nil {
			slog.Error("Rolling aggregation failed", "error", err)
		}

		slog.Info("Aggregation job completed", "duration", time.Since(startTime))
	}

	_, err := s.cron.AddFunc(s.config.AggregationSchedule, jobFunc)
	if err != nil {
		return fmt.Errorf("failed to add aggregation job: %w", err)
	}

	slog.Info("Registered aggregation job", "schedule", s.config.AggregationSchedule)
	return nil
}

// registerEmployeeSyncJob registers the employee sync job
func (s *Scheduler) registerEmployeeSyncJob() error {
	if s.employeeService == nil {
		return fmt.Errorf("employee service not configured")
	}

	jobFunc := func() {
		slog.Info("Starting scheduled employee sync job")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		startTime := time.Now()

		// TODO: Implement employee sync
		slog.InfoContext(ctx, "Employee sync job placeholder - implement when EmployeeService is ready")

		slog.Info("Employee sync job completed", "duration", time.Since(startTime))
	}

	_, err := s.cron.AddFunc(s.config.EmployeeSyncSchedule, jobFunc)
	if err != nil {
		return fmt.Errorf("failed to add employee sync job: %w", err)
	}

	slog.Info("Registered employee sync job", "schedule", s.config.EmployeeSyncSchedule)
	return nil
}

// registerGitActivitiesSyncJob registers the git activities sync job
func (s *Scheduler) registerGitActivitiesSyncJob() error {
	if s.gitSyncService == nil {
		return fmt.Errorf("git sync service not configured")
	}

	jobFunc := func() {
		slog.Info("Starting scheduled git activities sync job")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel()

		startTime := time.Now()

		// TODO: Implement git activities sync
		slog.InfoContext(ctx, "Git activities sync job placeholder - implement when GitSyncService is ready")

		slog.Info("Git activities sync job completed", "duration", time.Since(startTime))
	}

	_, err := s.cron.AddFunc(s.config.GitActivitiesSyncSchedule, jobFunc)
	if err != nil {
		return fmt.Errorf("failed to add git activities sync job: %w", err)
	}

	slog.Info("Registered git activities sync job", "schedule", s.config.GitActivitiesSyncSchedule)
	return nil
}

// TriggerAggregation manually triggers an aggregation job
func (s *Scheduler) TriggerAggregation(ctx context.Context) error {
	slog.Info("Manually triggering aggregation")

	if err := s.aggregationService.AggregateForCurrentMonth(ctx); err != nil {
		return fmt.Errorf("aggregation failed: %w", err)
	}

	if err := s.aggregationService.AggregateForLastNDays(ctx, 30); err != nil {
		slog.Warn("Rolling aggregation failed", "error", err)
	}

	return nil
}

// GetScheduledJobs returns information about scheduled jobs
func (s *Scheduler) GetScheduledJobs() []JobInfo {
	entries := s.cron.Entries()
	jobs := make([]JobInfo, len(entries))

	for i, entry := range entries {
		jobs[i] = JobInfo{
			ID:       int(entry.ID),
			Schedule: entry.Schedule.Next(time.Now()).Format(time.RFC3339),
			Next:     entry.Next,
			Prev:     entry.Prev,
		}
	}

	return jobs
}

// JobInfo represents information about a scheduled job
type JobInfo struct {
	ID       int       `json:"id"`
	Schedule string    `json:"schedule"`
	Next     time.Time `json:"next"`
	Prev     time.Time `json:"prev"`
}

// EmployeeService placeholder (will be implemented)
type EmployeeService struct{}

// GitSyncService placeholder (will be implemented)
type GitSyncService struct{}
