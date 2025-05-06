package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
)

const defaultRegion = "us-east-2"

// Address вернет адрес подключения к s3.
func (o *Options) Address() string {
	scheme := "http"
	if o.Secure {
		scheme = "https"
	}

	return scheme + "://" + o.Endpoint
}

// NewClient вернет клиента для s3 хранилища.
func NewClient(opts Options) (*s3.Client, error) {
	staticProvider := credentials.NewStaticCredentialsProvider(
		opts.AccessKeyID,
		opts.SecretAccessKey,
		"",
	)

	region := defaultRegion
	if opts.Region != "" {
		region = opts.Region
	}

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(staticProvider),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	cli := s3.NewFromConfig(
		cfg,
		func(options *s3.Options) {
			options.BaseEndpoint = aws.String(opts.Address())
			options.UsePathStyle = true
		},
	)

	return cli, nil
}
