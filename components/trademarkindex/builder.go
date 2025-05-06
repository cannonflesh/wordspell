package trademarkindex

import (
	"bytes"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/domain"
)

const sourceBatchLen = 5000

// DataSource извлекает все данные в память.
// Индекс, заполненный из каталога, нецелесообразно держать в памяти того же пода,
// где работает wordspell, у него свои актуальные индексы, и расход памяти будет неоправданно большим.
//
// Предлагаю запускать построение новых индексов по актуальным данным в отдельном поде
// (например, как Kubernetes CronJob, с некоторой периодичностью), записывать новые индексы в DataStore,
// а затем перезапускать поды приложений, использующих wordspell.
type DataSource interface {
	TradeMarkNames(int, int) ([]string, error)
}

type Builder struct {
	source DataSource
	store  DataStore
	logger *logrus.Entry
}

func NewBuilder(source DataSource, store DataStore, l *logrus.Entry) *Builder {
	return &Builder{
		source: source,
		store:  store,
		logger: l.WithField(domain.CategoryFieldName, "component.trademark_index_builder"),
	}
}

func (b *Builder) LoadIndexData() error {
	var res []string

	b.logger.Info("[TRADEMARK INDEX BUILD] start building")
	startBuild := time.Now()
	for start, read := 0, sourceBatchLen; read == sourceBatchLen; {
		names, err := b.source.TradeMarkNames(start, sourceBatchLen)
		if err != nil {
			return err
		}

		cleanupNames(names)
		res = append(res, names...)

		read = len(names)
		start += sourceBatchLen
	}
	b.logger.Infof("[TRADEMARK INDEX BUILD] trademarks total: %d, built in %v", len(res), time.Since(startBuild))

	startSave := time.Now()
	b.logger.Info("[TRADEMARK INDEX SAVE] start saving")
	err := b.store.Save(storeKey, bytes.NewBuffer([]byte(strings.Join(res, "\n"))))
	b.logger.Infof("[TRADEMARK INDEX SAVE] saved in %v", time.Since(startSave))

	return err
}

func cleanupNames(names []string) {
	for i := range names {
		names[i] = cleanupName(names[i])
	}
}

func cleanupName(name string) string {
	fields := strings.Fields(domain.CleanTextRE.ReplaceAllString(name, domain.SpaceSeparator))

	return strings.Join(fields, domain.SpaceSeparator)
}
