package integrationtest

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/maragudk/env"
)

func getAWSConfig() aws.Config {
	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolver(createAWSEndpointResolver()),
	)
	if err != nil {
		panic(err)
	}
	return awsConfig
}

func createAWSEndpointResolver() aws.EndpointResolverFunc {
	sqsEndpointURL := env.GetStringOrDefault("SQS_ENDPOINT_URL", "")
	if sqsEndpointURL == "" {
		panic("sqs endpoint URL must be set in testing with env var SQS_ENDPOINT_URL")
	}
	s3EndpointURL := env.GetStringOrDefault("S3_ENDPOINT_URL", "")
	if s3EndpointURL == "" {
		panic("s3 endpoint URL must be set in testing with env var S3_ENDPOINT_URL")
	}

	return func(service, region string) (aws.Endpoint, error) {
		switch service {
		case sqs.ServiceID:
			return aws.Endpoint{
				URL: sqsEndpointURL,
			}, nil
		case s3.ServiceID:
			return aws.Endpoint{
				URL: s3EndpointURL,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	}
}
