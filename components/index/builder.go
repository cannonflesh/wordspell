package index

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cannonflesh/wordspell/domain"
	"github.com/microcosm-cc/bluemonday"
)

var (
	htmlAddSpacesRE = regexp.MustCompile("([^\\s>])(<)")
)

const (
	categoryNameBatchLen = 10000
	itemDataBatchLen     = 100000
	pairFreqTreshold     = 50
	ruIndexFreqTreshold  = 23
	enIndexFreqTreshold  = 10
)

// DataSource извлекает все данные в память, но позволяет это делать постепенно.
// В конце концов результат приходится накапливать в памяти - но уже в "переваренном" виде,
// так он занимает много меньше места.
// В то же время, индекс, заполненный из каталога, нецелесообразно держать в памяти того же пода,
// где работает wordspell, у него свои актуальные индексы, и расход памяти будет неоправданно большим.
// Предлагаю запускать построение новых индексов по актуальным данным в отдельном поде
// (например, как Kubernetes CronJob, с некоторой периодичностью), записывать новые индексы в DataStore,
// а затем перезапускать поды приложений, использующих wordspell.
type DataSource interface {
	ItemData(start, limit int) ([]string, []string, error)
	CategoryNames(start, limit int) ([]string, error)
}

type Builder struct {
	source DataSource
	store  DataStore
	langs  langDetector
	logger *logrus.Entry
}

func NewBuilder(source DataSource, store DataStore, langs langDetector, l *logrus.Entry) *Builder {
	return &Builder{
		source: source,
		store:  store,
		langs:  langs,
		logger: l.WithField(domain.CategoryFieldName, "component.speller_index_builder"),
	}
}

func (b *Builder) LoadIndexFromDB() error {
	res := newData()

	err := b.buildItemIndex(res, itemDataBatchLen)
	if err != nil {
		return err
	}

	err = b.buildCategoryIndex(res, categoryNameBatchLen)
	if err != nil {
		return err
	}

	for k, v := range res.words[ruLangCode] {
		if v < ruIndexFreqTreshold {
			delete(res.words[ruLangCode], k)
		}
	}

	for k, v := range res.words[enLangCode] {
		if v < enIndexFreqTreshold {
			delete(res.words[enLangCode], k)
		}
	}

	for k, v := range res.dwords[ruLangCode] {
		if v < pairFreqTreshold {
			delete(res.dwords[ruLangCode], k)
		}
	}

	for k, v := range res.dwords[enLangCode] {
		if v < pairFreqTreshold {
			delete(res.dwords[enLangCode], k)
		}
	}

	res.words.merge(res.dwords)
	res.dwords = nil

	for lang := range res.words {
		startSave := time.Now()
		b.logger.Infof("[LANG INDEX SAVE] saving index, lang: %s", lang)
		if err = b.saveLangIndex(lang, res.words); err != nil {
			return err
		}
		b.logger.Infof("[LANG INDEX SAVE] index saved in %v, lang: %s", time.Since(startSave), lang)
	}

	return nil
}

func (b *Builder) buildItemIndex(res *data, batchSize int) error {
	start := 0
	startTime := time.Now()
	b.logger.Info("[ITEM INDEX BUILD] start building")

	totalNames := 0
	totalDescs := 0

	for {
		names, descs, err := b.source.ItemData(start, batchSize)
		if err != nil {
			return err
		}

		for _, n := range names {
			wordSlice := textPreProcess(n)
			b.processWordSlice(res, wordSlice)
		}

		for _, d := range descs {
			wordSlice := htmlPreProcess(d)
			b.processWordSlice(res, wordSlice)
		}

		totalNames += len(names)
		totalDescs += len(descs)
		start += batchSize

		b.logger.Infof(
			"[ITEM INDEX BUILD] total names: %d, total descriptions: %d, elapsed: %v",
			totalNames,
			totalDescs,
			time.Since(startTime),
		)

		if len(names) < batchSize {
			break
		}
	}

	b.logger.Infof(
		"[ITEM INDEX BUILD] all item data loaded in: %v",
		time.Since(startTime),
	)

	return nil
}

func (b *Builder) buildCategoryIndex(res *data, batchSize int) error {
	start := 0
	startTime := time.Now()
	b.logger.Info("[CATEGORY INDEX BUILD] start building")

	for {
		lines, err := b.source.CategoryNames(start, batchSize)
		if err != nil {
			return err
		}

		for _, l := range lines {
			wordSlice := textPreProcess(l)
			b.processWordSlice(res, wordSlice)
		}

		b.logger.Infof("[CATEGORY INDEX BUILD] total: %d, elapsed: %v", start+len(lines), time.Since(startTime))

		if len(lines) < batchSize {
			break
		}

		start += batchSize
	}

	b.logger.Infof(
		"[CATEGORY INDEX BUILD] all category names loaded in: %v",
		time.Since(startTime),
	)

	return nil
}

func (b *Builder) saveLangIndex(lang langCode, idx wordCollection) error {
	if _, found := idx[lang]; !found {
		return errors.New("no index for the such language: " + ruLangCode)
	}

	sorted := make([]*wordFrequency, 0, len(idx[lang]))
	for k, v := range idx[lang] {
		sorted = append(sorted, &wordFrequency{
			word:      k,
			frequency: v,
		})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].frequency > sorted[j].frequency
	})

	writeBuffer := new(bytes.Buffer)
	for _, line := range sorted {
		_, err := writeBuffer.Write(line.toLine())
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return b.store.Save(langCodeIndexKey(lang), writeBuffer)
}

const (
	wordPairSize = 2
	keySuffix    = ".index"
)

func (b *Builder) processWordSlice(d *data, ws []string) {
	for i := 0; i < len(ws); i++ {
		start := i
		end := i + wordPairSize
		if end > len(ws) {
			end = len(ws)
		}

		left, right, lang := b.langs.ParseWordPair(ws[start:end])
		if lang == unknownLangCode {
			continue
		}

		if left != "" {
			d.words[lang][left] = d.words[lang][left] + 1

			if right != "" {
				dword := left + " " + right
				d.dwords[lang][dword] = d.dwords[lang][dword] + 1
			}
		}
	}
}

func htmlPreProcess(in string) []string {
	p := bluemonday.StrictPolicy()

	line := strings.ToLower(
		domain.CleanIndexRE.ReplaceAllString(
			p.Sanitize(
				htmlAddSpacesRE.ReplaceAllString(
					in,
					"$1 $2",
				),
			),
			domain.SpaceSeparator,
		),
	)

	var res []string
	for _, w := range strings.Fields(line) {
		if wordFilter(w) {
			res = append(res, w)
		}
	}

	return res
}

func textPreProcess(in string) []string {
	line := strings.ToLower(domain.CleanIndexRE.ReplaceAllString(in, domain.SpaceSeparator))

	var res []string
	for _, w := range strings.Fields(line) {
		if wordFilter(w) {
			res = append(res, w)
		}
	}

	return res
}

func wordFilter(w string) bool {
	if runeLen(w) < 2 {
		return false
	}

	if strings.HasPrefix(w, "-") ||
		strings.HasSuffix(w, "-") ||
		strings.HasPrefix(w, "'") ||
		strings.HasPrefix(w, "`") {
		return false
	}

	return true
}

func langCodeIndexKey(l langCode) string {
	return l + keySuffix
}
