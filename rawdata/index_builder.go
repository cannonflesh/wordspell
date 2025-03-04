/* Здесь представлены средства загрузки индексов и фильтра из языковых файлов.
 * Сначала загружаются и сохраняются индексы (русский и английский), потом заполняется и сохраняется фильтр.
 * Точность работы спеллера сильно зависит от качества заполнения индекса.
 * Его можно "обучить", загрузив в него представительные и хорошо выверенные языковые файлы.
 *
 * В результате в директорию .datafiles будут сохранены три файла - русский и английский индексы
 * и общий фильтр Блума, содержащий всевозможные удаления по 1 и 2 руны из слов, содержащихся в индексах.
 *
 * В директории data находятся файлы, полученные загрузкой из Leipzig Corpora Collection - орфографических
 * коллекций общего назначения.
 */
package main

import (
	"github.com/cannonflesh/wordspell/components/bloomfilter"
	"github.com/cannonflesh/wordspell/components/wordmutate"
	"github.com/cannonflesh/wordspell/rawdata/wordfreq"
	"github.com/sirupsen/logrus"

	"github.com/cannonflesh/wordspell/components/index"
	"github.com/cannonflesh/wordspell/components/langdetect"
	"github.com/cannonflesh/wordspell/domain"
	"github.com/cannonflesh/wordspell/options"
)

func main() {
	logger := logrus.NewEntry(logrus.New())
	opt := &options.Options{DataDir: "./.datafiles"}

	// Реализованы два парсера, их описания - в файлах
	// rawtext/README.md и wordfreq/README.md соответственно.
	// В данный момент выбран wordfreq.Parser.
	// parser := rawtext.New(logger)
	parser := wordfreq.New(logger)

	lgr := logger.WithField(domain.CategoryFieldName, "utility.index_builder")

	////////// RUSSIAN INDEX ////////////
	indexRu, err := parser.BuildLangIndex(domain.RuLangCode)
	if err != nil {
		lgr.Fatal(err)
	}

	////////// ENGLISH INDEX ////////////
	indexEn, err := parser.BuildLangIndex(domain.EnLangCode)
	if err != nil {
		lgr.Fatal(err)
	}

	//////// Saving Indexes //////////
	langs := langdetect.New()
	idx := index.New(opt, langs, logger)
	idx.SetLangIndex(domain.RuLangCode, indexRu)
	idx.SetLangIndex(domain.EnLangCode, indexEn)
	err = idx.Save()
	if err != nil {
		lgr.Fatal(err)
	}

	//////// Filling Bloom Filter //////////
	bFilter := bloomfilter.New(opt, logger)
	mutator := wordmutate.New()
	fillBloomFilter(bFilter, mutator, idx, logger)
}

func fillBloomFilter(bFilter *bloomfilter.Component, mutate *wordmutate.Component, idx *index.Component, l *logrus.Entry) {
	logger := l.WithField(domain.CategoryFieldName, "utility.bloom_filter_filler")

	bFilterSizeRu, err := idx.DeletesEstimated(domain.RuLangCode)
	if err != nil {
		logger.Fatal(err)
	}
	bFilterSizeEn, err := idx.DeletesEstimated(domain.EnLangCode)
	if err != nil {
		logger.Fatal(err)
	}
	bFilter.Reset(bFilterSizeRu + bFilterSizeEn)

	ruWords, err := idx.Words(domain.RuLangCode)
	if err != nil {
		logger.Fatal(err)
	}

	enWords, err := idx.Words(domain.EnLangCode)
	if err != nil {
		logger.Fatal(err)
	}

	for w := range ruWords {
		dts := mutate.Deletes(w)
		bFilter.Add(dts...)
	}

	for w := range enWords {
		dts := mutate.Deletes(w)
		bFilter.Add(dts...)
	}

	err = bFilter.Save()
	if err != nil {
		logger.Fatal(err)
	}
}
