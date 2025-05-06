package main

import (
	"github.com/sirupsen/logrus"

	"github.com/cannonflesh/wordspell"
	"github.com/cannonflesh/wordspell/components/bloomfilter"
	"github.com/cannonflesh/wordspell/internal/postgres"
	s3client "github.com/cannonflesh/wordspell/internal/s3"
	"github.com/cannonflesh/wordspell/options"
	s3source "github.com/cannonflesh/wordspell/repo/s3"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	l := logrus.NewEntry(logger)

	opt := &options.Options{
		Bloom: bloomfilter.Options{FalsePositiveRate: 0.01},
		SiteDB: postgres.Options{
			Host:   "postgres",
			Port:   5432,
			DBName: "pgdb",
			User:   "user",
			Pass:   "pass",
		},
		S3Client: s3client.Options{
			Endpoint:        "minio:9000",
			AccessKeyID:     "minio",
			SecretAccessKey: "miniosecret",
		},
		S3Data: s3source.Options{
			Bucket: "wordspell-index",
			Name:   "wordspell",
		},
	}

	b, err := wordspell.NewBuilder(opt, l)
	if err != nil {
		l.Fatal(err)
	}

	err = b.Build()
	if err != nil {
		l.Fatal(err)
	}
}
