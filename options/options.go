package options

import (
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/components/bloomfilter"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/internal/postgres"
	s3client "gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/internal/s3"
	filerepo "gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/repo/file"
	s3repo "gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/repo/s3"
)

type Options struct {
	Bloom     bloomfilter.Options
	SiteDB    postgres.Options
	S3Client  s3client.Options
	S3Data    s3repo.Options
	DataFiles filerepo.Options
	Langs     []string
}
