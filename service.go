package wordspell

import (
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/processors/dimensions"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/processors/dimsuffix"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/processors/papersizes"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/processors/units"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/components/bloomfilter"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/components/index"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/components/langdetect"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/components/trademarkindex"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/components/wordmutate"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/domain"
	s3client "gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/internal/s3"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/options"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/processors/dupremove"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/processors/trademarks"
	s3source "gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/repo/s3"
)

type processor interface {
	Process(words []string) []string
}

type Service struct {
	langs  *langdetect.Component
	index  *index.Service
	mutate *wordmutate.Component
	bloom  *bloomfilter.Component

	preProcessors  []processor
	postProcessors []processor

	logger *logrus.Entry
}

func New(opt *options.Options, l *logrus.Entry) (*Service, error) {
	langDetect := langdetect.New()

	s3cli, err := s3client.NewClient(opt.S3Client)
	if err != nil {
		return nil, err
	}

	store, err := s3source.NewStore(s3cli, opt.S3Data)
	if err != nil {
		return nil, err
	}

	startTmLoad := time.Now()
	tm, err := trademarkindex.NewService(store, l)
	if err != nil {
		return nil, err
	}
	l.Infof("trademarks loaded in %s", time.Since(startTmLoad))

	startIdxLoad := time.Now()
	idx, err := index.NewService(opt, langDetect, store, l)
	if err != nil {
		return nil, err
	}
	l.Infof("index loaded in %s", time.Since(startIdxLoad))

	startLoadBloom := time.Now()
	bloom := bloomfilter.New(&opt.Bloom, store, l)
	err = bloom.Load()
	if err != nil {
		return nil, err
	}
	l.Infof("bloom loaded in %s", time.Since(startLoadBloom))

	preProcessors := []processor{
		trademarks.New(tm),
		dimsuffix.New(),
		dimensions.New(),
		papersizes.New(),
		units.New(),
	}

	postProcessors := []processor{
		dupremove.New(),
	}

	return &Service{
		langs:  langDetect,
		index:  idx,
		mutate: wordmutate.New(),
		bloom:  bloom,

		preProcessors:  preProcessors,
		postProcessors: postProcessors,

		logger: l.WithField(domain.CategoryFieldName, "service.word_speller"),
	}, nil
}

func (s *Service) Correct(request string) string {
	preProcessed := strings.Fields(domain.CleanTextRE.ReplaceAllString(request, domain.SpaceSeparator))
	for _, wp := range s.preProcessors {
		preProcessed = wp.Process(preProcessed)
	}

	digest := s.checkWordPairs(domain.ParseDigest(preProcessed))

	res := make([]string, 0, len(digest))
	for _, v := range digest {
		switch vv := v.(type) {
		case domain.DigestRaw:
			if splitted, found := s.splittedWord(vv); found {
				res = append(res, splitted)
			} else {
				res = append(res, s.correctWord(vv).String())
			}
		case domain.DigestReady:
			res = append(res, vv.String())
		}
	}

	for _, wp := range s.postProcessors {
		res = wp.Process(res)
	}

	return strings.Join(res, domain.SpaceSeparator)
}

func (s *Service) checkWordPairs(dig domain.Digest) domain.Digest {
	res := domain.NewEmptyDigest()

	for el, replaced := s.wordPair(dig); el != nil; el, replaced = s.wordPair(dig) {
		res = res.Add(el)

		if len(dig) == 0 {
			break
		}

		if replaced {
			dig = dig[2:]
		} else {
			dig = dig[1:]
		}
	}

	return res
}

func (s *Service) wordPair(dig domain.Digest) (domain.DigestElement, bool) {
	if len(dig) == 0 {
		return nil, false
	}

	var (
		left, right         domain.DigestRaw
		okLeft, okRight     bool
		leftLang, rightLang string
	)

	if left, okLeft = dig[0].(domain.DigestRaw); !okLeft {
		return dig[0], false
	}

	if len(dig) < 2 {
		return left, false
	}

	if right, okRight = dig[1].(domain.DigestRaw); !okRight {
		return left, false
	}

	leftLang = s.langs.LangByWord(left.String())
	rightLang = s.langs.LangByWord(right.String())

	if leftLang == domain.UnknownLangCode || leftLang == domain.NumLangCode || rightLang != leftLang {
		return left, false
	}

	merged := strings.ToLower(left.Merge(right).String())
	if s.index.Weight(merged) > 0 {
		return domain.NewDigestReady(merged), true
	}

	return left, false
}

func (s *Service) splittedWord(el domain.DigestRaw) (string, bool) {
	splitted := s.mutate.InsertSpace(strings.ToLower(el.String()))
	var (
		maxWeight uint32
		best      string
	)

	for _, w := range splitted {
		if weight := s.index.Weight(w); weight > maxWeight {
			maxWeight = weight
			best = w
		}
	}

	if best != "" {
		return best, true
	}

	return el.String(), false
}

func (s *Service) correctWord(el domain.DigestRaw) domain.DigestElement {
	word := strings.ToLower(el.String())

	if s.index.Weight(word) > 0 {
		return domain.NewDigestReady(word)
	}

	dels := s.mutate.Deletes(word)
	for _, w := range dels {
		// Проверяем, нет ли в индексе самого удаления.
		if weight := s.index.Weight(w); weight > 0 {
			return domain.NewDigestReady(w)
		}

		if s.bloom.Test(w) {
			// Выполняем полный набор вставок по одной руне, проверяем на наличие их в индексе.
			insertsOne := s.insertRune(w)
			correctWord := s.findWordWithMaxWeight(insertsOne)
			if correctWord != "" {
				return domain.NewDigestReady(correctWord)
			}

			// Для каждой из однорунных вставок выполняем полный набор однорунных вставок
			// и проверяем еще и их на наличие в индексе.
			for _, plusOne := range insertsOne {
				correctWord = s.findWordWithMaxWeight(s.insertRune(plusOne))
				if correctWord != "" {
					return domain.NewDigestReady(correctWord)
				}
			}
		}
	}

	return el
}

func (s *Service) insertRune(w string) []string {
	lang := s.langs.LangByWord(w)
	switch lang {
	case domain.RuLangCode:
		return s.mutate.InsertRuneRu(w)
	case domain.EnLangCode:
		return s.mutate.InsertRuneEn(w)
	case domain.NumLangCode:
		return []string{w}
	}

	s.logger.Debug("correctWord: language not detected")

	return nil
}

func (s *Service) findWordWithMaxWeight(words []string) string {
	maxWeight := uint32(0)
	res := ""

	for _, w := range words {
		if weight := s.index.Weight(w); weight > maxWeight {
			maxWeight = weight
			res = w
		}
	}

	return res
}
