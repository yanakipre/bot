package scheduletooling

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	quartzlogger "github.com/reugn/go-quartz/logger"
	"github.com/reugn/go-quartz/quartz"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/clouderr"
	"github.com/yanakipre/bot/internal/logger"
)

// Scheduler is a wrapper around quartz.Scheduler.
// It knows that Job brings fresh configs with it,
// and enables/reschedules Job when config changes.
type Scheduler struct {
	// collectedJobsToSchedule is a buffer to be filled with jobs to add right after scheduler start.
	// this is to address the issue, when jobs can be added only after scheduler start,
	// https://github.com/reugn/go-quartz/issues/50
	collectedJobsToSchedule []Job
	jobUpdateInterval       time.Duration
	jobsMutex               *sync.Mutex
	jobs                    map[string]Job
	scheduler               quartz.Scheduler
	configReapplyInterval   time.Duration //nolint:unused,structcheck
	started                 *sync.Once
	isRunning               bool
}

// Add silently adds job to scheduler or ignores it
func (s *Scheduler) Add(ctx context.Context, job Job) error {
	if s.isRunning {
		// we error out because it's easier.
		// If one wishes to add jobs after scheduler start (e.g. at runtime),
		// reuse scheduleCollectedJobs here and add a test.
		return fmt.Errorf("already running, can't add job")
	}
	cfg, err := job.GetConfig()
	if err != nil {
		logger.Error(
			ctx,
			clouderr.WrapWithFields(
				fmt.Errorf("could not get config: %w", err),
				jobLogKeys(job)...,
			),
		)
		return err
	}
	if !cfg.Enabled {
		logger.Info(ctx, "job is disabled", jobLogKeys(job)...)
		return nil
	}
	s.collectedJobsToSchedule = append(s.collectedJobsToSchedule, job)
	return nil
}

// Add silently adds job to scheduler or ignores it
func (s *Scheduler) scheduleCollectedJobs(ctx context.Context) {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	for i := range s.collectedJobsToSchedule {
		s.scheduleSilently(ctx, s.collectedJobsToSchedule[i])
	}
}

func jobLogKeys(j Job, fields ...zap.Field) []zap.Field {
	return append(fields, zap.String("job_description", j.Description()))
}

// scheduleSilently starts scheduling the job
func (s *Scheduler) scheduleSilently(ctx context.Context, job Job) {
	s.stopScheduling(ctx, job)

	s.jobs[job.Key().String()] = job

	if err := s.scheduler.ScheduleJob(quartz.NewJobDetail(job, job.Key()), job); err != nil {
		logger.Error(
			ctx,
			clouderr.WrapWithFields(
				fmt.Errorf("could not schedule job, skipping: %w", err),
				jobLogKeys(job)...,
			),
		)
		return
	}
}

// stopScheduling removes job from scheduling, so it's not executed anymore
func (s *Scheduler) stopScheduling(ctx context.Context, job Job) {
	if err := s.scheduler.DeleteJob(job.Key()); err != nil {
		// The following happens during application start,
		// when we try to delete a job before scheduling it first time.
		//
		// That behavior is intentional, because we do the rescheduling,
		// when live config changes without application restart: we want to delete
		// the existing job, then schedule it again with new config.
		if errors.Is(err, quartz.ErrJobNotFound) {
			// we allow deleting non-existent jobs.
			return
		}
		logger.Error(
			ctx,
			clouderr.WrapWithFields(
				fmt.Errorf("could not stop job, skipping: %w", err),
				jobLogKeys(job)...,
			),
		)
	}
}

// Start scheduling.
func (s *Scheduler) Start(ctx context.Context) {
	s.started.Do(func() {
		s.isRunning = true
		rescheduler := s.updateSchedules(logger.WithName(ctx, "scheduler"))
		go rescheduler(logger.WithName(ctx, "scheduler"))
		s.scheduler.Start(ctx)
		s.scheduleCollectedJobs(ctx)
	})
}

// Stop scheduling.
func (s *Scheduler) Stop() { s.scheduler.Stop() }

func (s *Scheduler) Wait(ctx context.Context) {
	s.scheduler.Wait(ctx)
	//
	// Close all the jobs that want to free resources.
	//
	wg := &sync.WaitGroup{}

	ctx = logger.WithName(ctx, "closer")

	s.jobsMutex.Lock()
	for _, job := range s.jobs {
		wg.Add(1)
		go func(job func()) {
			defer wg.Done()
			job()
		}(func() {
			if err := job.Close(ctx); err != nil {
				logger.Error(
					ctx,
					clouderr.WrapWithFields(
						fmt.Errorf("could not close job: %w", err),
						jobLogKeys(job)...,
					),
				)
			}
			logger.Info(ctx, "closed job", zap.String("job_description", job.Description()))
		})
	}
	s.jobsMutex.Unlock()

	sig := make(chan struct{})
	go func() { defer close(sig); wg.Wait() }()

	select {
	case <-ctx.Done():
		logger.Info(ctx, "did not make it in time to close jobs")
	case <-sig:
		logger.Info(ctx, "closed all jobs")
	}
}

// quartzBuiltinLogger makes built-in quartz logger comply to our .
// Based on example from go-quartz tests.
// Prevents lines like:
// WARN 2024/10/31 12:11:00 scheduler.go:538: Job default::transactional_outbox terminated with error: ... omitted
func quartzBuiltinLogger() {
	quartzlogger.SetDefault(&quartzLog{
		l: logger.FromContext(context.Background()).Named("quartz"),
	})
}

func NewScheduler(jobUpdateInterval time.Duration) *Scheduler {
	quartzBuiltinLogger()
	return &Scheduler{
		started:           &sync.Once{},
		jobs:              map[string]Job{},
		jobUpdateInterval: jobUpdateInterval,
		scheduler:         quartz.NewStdScheduler(),
		jobsMutex:         &sync.Mutex{},
	}
}

func init() {
	log.SetOutput(io.Discard)
}
