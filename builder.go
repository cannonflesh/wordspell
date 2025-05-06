package wordspell

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/cannonflesh/wordspell/components/bloomfilter"
	"github.com/cannonflesh/wordspell/components/index"
	"github.com/cannonflesh/wordspell/components/langdetect"
	"github.com/cannonflesh/wordspell/components/trademarkindex"
	"github.com/cannonflesh/wordspell/components/wordmutate"
	"github.com/cannonflesh/wordspell/domain"
	"github.com/cannonflesh/wordspell/internal/postgres"
	"github.com/cannonflesh/wordspell/internal/s3"
	"github.com/cannonflesh/wordspell/options"
	"github.com/cannonflesh/wordspell/repo/catalog"
	s3repo "github.com/cannonflesh/wordspell/repo/s3"
)

type Builder struct {
	opt *options.Options

	indexBuilder          *index.Builder
	tradeMarkIndexBuilder *trademarkindex.Builder

	store  index.DataStore
	logger *logrus.Entry
}

func NewBuilder(opt *options.Options, l *logrus.Entry) (*Builder, error) {
	lang := langdetect.New()

	pgConn, err := postgres.New(&opt.SiteDB)
	if err != nil {
		return nil, err
	}
	source := catalog.New(pgConn, l)

	s3cli, err := s3.NewClient(opt.S3Client)
	if err != nil {
		return nil, err
	}
	store, err := s3repo.NewStore(s3cli, opt.S3Data)
	if err != nil {
		return nil, err
	}

	return &Builder{
		opt: opt,

		indexBuilder:          index.NewBuilder(source, store, lang, l),
		tradeMarkIndexBuilder: trademarkindex.NewBuilder(source, store, l),

		store:  store,
		logger: l.WithField(domain.CategoryFieldName, "service.indexes_builder"),
	}, nil
}

func (b *Builder) Build() error {
	err := b.indexBuilder.LoadIndexFromDB()
	if err != nil {
		return err
	}

	err = b.tradeMarkIndexBuilder.LoadIndexData()
	if err != nil {
		return err
	}

	idx, err := index.NewService(b.opt, langdetect.New(), b.store, b.logger)
	if err != nil {
		return err
	}

	bloom := bloomfilter.New(&b.opt.Bloom, b.store, b.logger)

	b.logger.Info("[BLOOM FILTER BUILD] start building")
	startBloomBuild := time.Now()
	err = fillBloomFilter(bloom, idx, wordmutate.New())
	if err != nil {
		return err
	}
	b.logger.Infof("[BLOOM FILTER BUILD] built in %v", time.Since(startBloomBuild))

	b.logger.Info("[BLOOM FILTER SAVE] start saving")
	startBloomSave := time.Now()
	err = bloom.Save()
	if err != nil {
		return err
	}
	b.logger.Infof("[BLOOM FILTER SAVE] saved in %v", time.Since(startBloomSave))

	return nil
}

func fillBloomFilter(bFilter *bloomfilter.Component, idx *index.Service, mutate *wordmutate.Component) error {
	bFilterSize, err := idx.DeletesEstimated()
	if err != nil {
		return err
	}

	bFilter.Reset(bFilterSize)
	if err != nil {
		return err
	}

	idxWords, err := idx.Words()
	if err != nil {
		return err
	}

	for w := range idxWords {
		dts := mutate.Deletes(w)
		bFilter.Add(dts...)
	}

	return nil
}
