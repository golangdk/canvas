package integrationtest

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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
	return func(service, region string) (aws.Endpoint, error) {
		if sqsEndpointURL != "" && service == sqs.ServiceID {
			return aws.Endpoint{
				URL: sqsEndpointURL,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	}
}
