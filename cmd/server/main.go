// Package main is the entry point to the server. It reads configuration, sets up logging and error handling,
// handles signals from the OS, and starts and stops the server.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/smithy-go/logging"
	"github.com/maragudk/env"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"canvas/jobs"
	"canvas/messaging"
	"canvas/server"
	"canvas/storage"
)

// release is set through the linker at build time, generally from a git sha.
// Used for logging and error reporting.
var release string

func main() {
	os.Exit(start())
}

func start() int {
	_ = env.Load()

	logEnv := env.GetStringOrDefault("LOG_ENV", "development")
	log, err := createLogger(logEnv)
	if err != nil {
		fmt.Println("Error setting up the logger:", err)
		return 1
	}
	log = log.With(zap.String("release", release))
	defer func() {
		// If we cannot sync, there's probably something wrong with outputting logs,
		// so we probably cannot write using fmt.Println either. So just ignore the error.
		_ = log.Sync()
	}()

	host := env.GetStringOrDefault("HOST", "localhost")
	port := env.GetIntOrDefault("PORT", 8080)

	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithLogger(createAWSLogAdapter(log)),
		config.WithEndpointResolver(createAWSEndpointResolver()),
	)
	if err != nil {
		log.Info("Error creating AWS config", zap.Error(err))
		return 1
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())

	queue := createQueue(log, awsConfig)
	db := createDatabase(log, registry)
	if err := db.Connect(); err != nil {
		log.Info("Error connecting to database", zap.Error(err))
		return 1
	}

	s := server.New(server.Options{
		AdminPassword:   env.GetStringOrDefault("ADMIN_PASSWORD", "eyDawVH9LLZtaG2q"),
		Database:        db,
		Host:            host,
		Log:             log,
		MetricsPassword: env.GetStringOrDefault("METRICS_PASSWORD", "12345678"),
		Metrics:         registry,
		Port:            port,
		Queue:           queue,
	})

	r := jobs.NewRunner(jobs.NewRunnerOptions{
		BlobStore: createBlobStore(log, awsConfig),
		Emailer:   createEmailer(log, host, port),
		Log:       log,
		Metrics:   registry,
		Queue:     queue,
	})

	var eg errgroup.Group
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	eg.Go(func() error {
		<-ctx.Done()
		if err := s.Stop(); err != nil {
			log.Info("Error stopping server", zap.Error(err))
			return err
		}
		return nil
	})

	eg.Go(func() error {
		r.Start(ctx)
		return nil
	})

	if err := s.Start(); err != nil {
		log.Info("Error starting server", zap.Error(err))
		return 1
	}

	if err := eg.Wait(); err != nil {
		return 1
	}
	return 0
}

func createLogger(env string) (*zap.Logger, error) {
	switch env {
	case "production":
		return zap.NewProduction()
	case "development":
		return zap.NewDevelopment()
	default:
		return zap.NewNop(), nil
	}
}

func createAWSLogAdapter(log *zap.Logger) logging.LoggerFunc {
	return func(classification logging.Classification, format string, v ...interface{}) {
		switch classification {
		case logging.Debug:
			log.Sugar().Debugf(format, v...)
		case logging.Warn:
			log.Sugar().Warnf(format, v...)
		}
	}
}

// createAWSEndpointResolver used for local development endpoints.
// See https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/endpoints/
func createAWSEndpointResolver() aws.EndpointResolverFunc {
	sqsEndpointURL := env.GetStringOrDefault("SQS_ENDPOINT_URL", "")
	s3EndpointURL := env.GetStringOrDefault("S3_ENDPOINT_URL", "")

	return func(service, region string) (aws.Endpoint, error) {
		switch service {
		case sqs.ServiceID:
			if sqsEndpointURL != "" {
				return aws.Endpoint{
					URL: sqsEndpointURL,
				}, nil
			}
		case s3.ServiceID:
			if s3EndpointURL != "" {
				return aws.Endpoint{
					URL: s3EndpointURL,
				}, nil
			}
		}
		// Fallback to default endpoint
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	}
}

func createDatabase(log *zap.Logger, registry *prometheus.Registry) *storage.Database {
	return storage.NewDatabase(storage.NewDatabaseOptions{
		Host:                  env.GetStringOrDefault("DB_HOST", "localhost"),
		Port:                  env.GetIntOrDefault("DB_PORT", 5432),
		User:                  env.GetStringOrDefault("DB_USER", ""),
		Password:              env.GetStringOrDefault("DB_PASSWORD", ""),
		Name:                  env.GetStringOrDefault("DB_NAME", ""),
		MaxOpenConnections:    env.GetIntOrDefault("DB_MAX_OPEN_CONNECTIONS", 10),
		MaxIdleConnections:    env.GetIntOrDefault("DB_MAX_IDLE_CONNECTIONS", 10),
		ConnectionMaxLifetime: env.GetDurationOrDefault("DB_CONNECTION_MAX_LIFETIME", time.Hour),
		Log:                   log,
		Metrics:               registry,
	})
}

func createQueue(log *zap.Logger, awsConfig aws.Config) *messaging.Queue {
	return messaging.NewQueue(messaging.NewQueueOptions{
		Config:   awsConfig,
		Log:      log,
		Name:     env.GetStringOrDefault("QUEUE_NAME", "jobs"),
		WaitTime: env.GetDurationOrDefault("QUEUE_WAIT_TIME", 20*time.Second),
	})
}

func createEmailer(log *zap.Logger, host string, port int) *messaging.Emailer {
	return messaging.NewEmailer(messaging.NewEmailerOptions{
		BaseURL:            env.GetStringOrDefault("BASE_URL", fmt.Sprintf("http://%v:%v", host, port)),
		Log:                log,
		MarketingEmailName: env.GetStringOrDefault("MARKETING_EMAIL_NAME", "Canvas bot"),
		MarketingEmailAddress: env.GetStringOrDefault("MARKETING_EMAIL_ADDRESS",
			"bot@marketing.example.com"),
		Token:                  env.GetStringOrDefault("POSTMARK_TOKEN", ""),
		TransactionalEmailName: env.GetStringOrDefault("TRANSACTIONAL_EMAIL_NAME", "Canvas bot"),
		TransactionalEmailAddress: env.GetStringOrDefault("TRANSACTIONAL_EMAIL_ADDRESS",
			"bot@transactional.example.com"),
	})
}

func createBlobStore(log *zap.Logger, awsConfig aws.Config) *storage.BlobStore {
	return storage.NewBlobStore(storage.NewBlobStoreOptions{
		Bucket:    env.GetStringOrDefault("BUCKET_NAME", "assets"),
		Config:    awsConfig,
		Log:       log,
		PathStyle: env.GetBoolOrDefault("BLOB_STORE_PATH_STYLE", false),
	})
}
