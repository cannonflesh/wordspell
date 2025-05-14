package options

import (
	"github.com/cannonflesh/wordspell/components/bloomfilter"
	"github.com/cannonflesh/wordspell/internal/postgres"
	s3client "github.com/cannonflesh/wordspell/internal/s3"
	s3repo "github.com/cannonflesh/wordspell/repo/s3"
)

type Options struct {
	Bloom    bloomfilter.Options
	SiteDB   postgres.Options
	S3Client s3client.Options
	S3Data   s3repo.Options
	Langs    []string
}
