// Package jobs has a Runner that can run registered jobs in parallel.
package jobs

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	"canvas/messaging"
	"canvas/model"
)

// Runner runs jobs.
type Runner struct {
	emailer        *messaging.Emailer
	jobCount       *prometheus.CounterVec
	jobDurations   *prometheus.CounterVec
	jobs           map[string]Func
	log            *zap.Logger
	queue          *messaging.Queue
	runnerReceives *prometheus.CounterVec
}

type NewRunnerOptions struct {
	Emailer *messaging.Emailer
	Log     *zap.Logger
	Metrics *prometheus.Registry
	Queue   *messaging.Queue
}

func NewRunner(opts NewRunnerOptions) *Runner {
	if opts.Log == nil {
		opts.Log = zap.NewNop()
	}

	if opts.Metrics == nil {
		opts.Metrics = prometheus.NewRegistry()
	}

	jobCount := promauto.With(opts.Metrics).NewCounterVec(prometheus.CounterOpts{
		Name: "app_jobs_total",
	}, []string{"name", "success"})

	jobDurations := promauto.With(opts.Metrics).NewCounterVec(prometheus.CounterOpts{
		Name: "app_job_duration_seconds_total",
	}, []string{"name", "success"})

	runnerReceives := promauto.With(opts.Metrics).NewCounterVec(prometheus.CounterOpts{
		Name: "app_job_runner_receives_total",
	}, []string{"success"})

	return &Runner{
		emailer:        opts.Emailer,
		jobCount:       jobCount,
		jobDurations:   jobDurations,
		jobs:           map[string]Func{},
		log:            opts.Log,
		queue:          opts.Queue,
		runnerReceives: runnerReceives,
	}
}

// Func is the actual work to do in a job.
// The given context is the root context of the runner, which may be cancelled.
type Func = func(context.Context, model.Message) error

// Start the Runner, blocking until the given context is cancelled.
func (r *Runner) Start(ctx context.Context) {
	r.log.Info("Starting")
	r.registerJobs()
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			r.log.Info("Stopping")
			wg.Wait()
			return
		default:
			r.receiveAndRun(ctx, &wg)
		}
	}
}

// receiveAndRun jobs.
func (r *Runner) receiveAndRun(ctx context.Context, wg *sync.WaitGroup) {
	m, receiptID, err := r.queue.Receive(ctx)
	if err != nil {
		r.runnerReceives.WithLabelValues("false").Inc()
		r.log.Info("Error receiving message", zap.Error(err))
		// Sleep a bit to not hammer the queue if there's an error with it
		time.Sleep(time.Second)
		return
	}

	// If there was no message there is nothing to do
	if m == nil {
		r.runnerReceives.WithLabelValues("true").Inc()
		return
	}

	name, ok := (*m)["job"]
	if !ok {
		r.runnerReceives.WithLabelValues("false").Inc()
		r.log.Info("Error getting job name from message")
		return
	}

	job, ok := r.jobs[name]
	if !ok {
		r.runnerReceives.WithLabelValues("false").Inc()
		r.log.Info("No job with this name", zap.String("name", name))
		return
	}

	r.runnerReceives.WithLabelValues("true").Inc()

	wg.Add(1)
	go func() {
		defer wg.Done()

		log := r.log.With(zap.String("name", name))

		defer func() {
			if rec := recover(); rec != nil {
				r.jobCount.WithLabelValues(name, "false").Inc()
				log.Info("Recovered from panic in job", zap.Any("recover", rec))
			}
		}()

		before := time.Now()
		err := job(ctx, *m)
		duration := time.Since(before)

		success := strconv.FormatBool(err == nil)
		r.jobCount.WithLabelValues(name, success).Inc()
		r.jobDurations.WithLabelValues(name, success).Add(duration.Seconds())

		if err != nil {
			log.Info("Error running job", zap.Error(err))
			return
		}
		log.Info("Successfully ran job", zap.Duration("duration", duration))

		// We use context.Background as the parent context instead of the existing ctx, because if we've come
		// this far we don't want the deletion to be cancelled.
		deleteCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := r.queue.Delete(deleteCtx, receiptID); err != nil {
			log.Info("Error deleting message, job will be repeated", zap.Error(err))
		}
	}()
}

// registry provides a way to Register jobs by name.
type registry interface {
	Register(name string, fn Func)
}

// Register implements registry.
func (r *Runner) Register(name string, j Func) {
	r.jobs[name] = j
}
