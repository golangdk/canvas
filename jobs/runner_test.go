package jobs_test

import (
	"context"
	"testing"

	"github.com/matryer/is"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"canvas/integrationtest"
	"canvas/jobs"
	"canvas/model"
)

type testRegistry map[string]jobs.Func

func (r testRegistry) Register(name string, fn jobs.Func) {
	r[name] = fn
}

func TestRunner_Start(t *testing.T) {
	integrationtest.SkipIfShort(t)

	t.Run("starts the runner and runs jobs until the context is cancelled", func(t *testing.T) {
		is := is.New(t)

		queue, cleanup := integrationtest.CreateQueue()
		defer cleanup()

		log, logs := newLogger()

		runner := jobs.NewRunner(jobs.NewRunnerOptions{
			Log:   log,
			Queue: queue,
		})

		ctx, cancel := context.WithCancel(context.Background())

		runner.Register("test", func(ctx context.Context, m model.Message) error {
			foo, ok := m["foo"]
			is.True(ok)
			is.Equal("bar", foo)

			cancel()
			return nil
		})

		err := queue.Send(context.Background(), model.Message{"job": "test", "foo": "bar"})
		is.NoErr(err)

		// This blocks until the context is cancelled by the job function
		runner.Start(ctx)

		is.Equal(3, logs.Len())
		is.Equal("Starting", logs.All()[0].Message)
		is.Equal("Successfully ran job", logs.All()[1].Message)
		is.Equal("Stopping", logs.All()[2].Message)
	})

	t.Run("emits job metrics", func(t *testing.T) {
		is := is.New(t)

		queue, cleanup := integrationtest.CreateQueue()
		defer cleanup()

		registry := prometheus.NewRegistry()

		runner := jobs.NewRunner(jobs.NewRunnerOptions{
			Metrics: registry,
			Queue:   queue,
		})

		ctx, cancel := context.WithCancel(context.Background())

		runner.Register("test", func(ctx context.Context, m model.Message) error {
			cancel()
			return nil
		})

		err := queue.Send(context.Background(), model.Message{"job": "test"})
		is.NoErr(err)

		runner.Start(ctx)

		metrics, err := registry.Gather()
		is.NoErr(err)
		is.Equal(3, len(metrics))

		metric := metrics[0]
		is.Equal("app_job_duration_seconds_total", metric.GetName())
		is.Equal("name", metric.Metric[0].Label[0].GetName())
		is.Equal("test", metric.Metric[0].Label[0].GetValue())
		is.Equal("success", metric.Metric[0].Label[1].GetName())
		is.Equal("true", metric.Metric[0].Label[1].GetValue())
		is.True(metric.Metric[0].Counter.GetValue() > 0)

		metric = metrics[1]
		is.Equal("app_job_runner_receives_total", metric.GetName())
		is.Equal("success", metric.Metric[0].Label[0].GetName())
		is.Equal("true", metric.Metric[0].Label[0].GetValue())
		is.True(metric.Metric[0].Counter.GetValue() > 0)

		metric = metrics[2]
		is.Equal("app_jobs_total", metric.GetName())
		is.Equal("name", metric.Metric[0].Label[0].GetName())
		is.Equal("test", metric.Metric[0].Label[0].GetValue())
		is.Equal("success", metric.Metric[0].Label[1].GetName())
		is.Equal("true", metric.Metric[0].Label[1].GetValue())
		is.Equal(float64(1), metric.Metric[0].Counter.GetValue())
	})
}

func newLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.InfoLevel)
	return zap.New(core), logs
}
