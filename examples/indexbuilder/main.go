package main

import (
	"github.com/sirupsen/logrus"

	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/components/bloomfilter"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/internal/postgres"
	s3client "gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/internal/s3"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/options"
	s3source "gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/repo/s3"
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
