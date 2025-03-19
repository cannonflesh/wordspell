package wordspell

import (
	"github.com/sirupsen/logrus"

	"github.com/cannonflesh/wordspell/components/bloomfilter"
	"github.com/cannonflesh/wordspell/components/index"
	"github.com/cannonflesh/wordspell/components/langdetect"
	"github.com/cannonflesh/wordspell/components/wordmutate"
	"github.com/cannonflesh/wordspell/domain"
	"github.com/cannonflesh/wordspell/options"
)

type Service struct {
	opt *options.Options

	langs  *langdetect.Component
	index  *index.Component
	mutate *wordmutate.Component
	bloom  *bloomfilter.Component

	logger *logrus.Entry
}

func New(opt *options.Options, l *logrus.Entry) *Service {
	langDetect := langdetect.New()

	return &Service{
		opt: opt,

		langs:  langDetect,
		index:  index.New(opt, langDetect, l),
		mutate: wordmutate.New(),
		bloom:  bloomfilter.New(opt, l),
		logger: l.WithField(domain.CategoryFieldName, "service.word_speller"),
	}
}

func (s *Service) Correct(word string) string {
	if s.index.Weight(word) > 0 {
		return word
	}

	dels := s.mutate.Deletes(word)
	maxWeight := uint32(0)
	best := ""

	for _, w := range dels {
		// Проверяем, нет ли в индексе самого удаления.
		if weight := s.index.Weight(w); weight > 0 {
			best = w

			break
		}

		if s.bloom.Test(w) {
			// Выполняем полный набор вставок по одной руне, проверяем на наличие их в индексе.
			insertsOne := s.insertRune(w)
			for _, plusOne := range insertsOne {
				if weight := s.index.Weight(plusOne); weight > maxWeight {
					maxWeight = weight
					best = plusOne
				}
			}
			if best != "" {
				break
			}

			// Для каждой из однорунных вставок выполняем полный набор однорунных вставок
			// и проверяем еще и их на наличие в индексе.
			for _, plusOne := range insertsOne {
				insertsTwo := s.insertRune(plusOne)
				for _, plusTwo := range insertsTwo {
					if weight := s.index.Weight(plusTwo); weight > maxWeight {
						maxWeight = weight
						best = plusTwo
					}

					if best != "" {
						break
					}
				}
			}
		}
	}

	if best == "" {
		best = word
	}

	return best
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

	s.logger.Warn("Correct: language not detected")

	return nil
}
