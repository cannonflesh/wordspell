package bloomfilter

import (
	"bytes"
	"io"
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cannonflesh/wordspell/domain"
)

const (
	defaultFalsePositiveRate = 0.005
	defaultFilterSize        = 10000
	storeKey                 = "bloom.dat"
)

type DataStore interface {
	DataReader(key string) (io.ReadCloser, error)
	IsExist(key string) (bool, error)
	Save(key string, content io.Reader) error
}

type Component struct {
	falsePositiveRate float64
	impl              *bloom.BloomFilter
	store             DataStore
	logger            *logrus.Entry

	mu sync.Mutex
}

// New создает пустой filter, который на любой тест отвечает false.
// Чтобы сделать его полезным, нужно сделать f.Reset(11223344),
// где параметр равен количеству добавленных к фильтру элементов.
// И потом выполнить f.Add(el...) для полного списка элементов, хранимых в фильтре.
//
// Тогда фильтр всегда ответит f.Test(el) == true для элемента,
// который в него добавляли, и с вероятностью f.falsePositiveRate ответит false
// для элемента, который не был в него добавлен.
func New(opt *Options, store DataStore, logger *logrus.Entry) *Component {
	fpr := defaultFalsePositiveRate
	if opt.FalsePositiveRate > 0.0 {
		fpr = opt.FalsePositiveRate
	}

	res := &Component{
		falsePositiveRate: fpr,
		impl:              bloom.NewWithEstimates(defaultFilterSize, fpr),
		store:             store,
		logger:            logger.WithField(domain.CategoryFieldName, "components.bloom_filter"),
	}

	return res
}

// Reset инициирует фильтр, после этой процедуры его можно начинать заполнять.
// принимает точное количество элементов, которые будут содержаться в фильтре.
// Если указать значение меньше, вероятность ложноположительных тестов
// будет выше f.falsePositiveRate.
// Если же указать значение больше, вероятность будет гарантирована,
// но память, занимаемая фильтром, будет больше оптимальной.
func (c *Component) Reset(size uint) {
	c.impl = bloom.NewWithEstimates(size, c.falsePositiveRate)
}

// Add добавляет элементы в фильтр. Элементы - это строки любого размера.
func (c *Component) Add(words ...string) {
	if c.impl == nil {
		c.logger.Warn("bloom filter not initialized, nothing changed")

		return
	}

	for _, w := range words {
		_ = c.impl.Add([]byte(w))
	}
}

// Test - рабочий метод фильтра.
// - Всегда возвращает true для элементов, ранее добавленных к фильтру.
// - С вероятностью falsePositiveRate возвращает false для элементов, которые не добавляли к фильтру.
func (c *Component) Test(w string) bool {
	if c.impl == nil {
		c.logger.Warn("not initialized bloom filter always returnt false")

		return false
	}

	return c.impl.TestString(w)
}

// Save - записывает заполненный фильтр в DataStore.
func (c *Component) Save() error {
	data, err := c.impl.GobEncode()
	if err != nil {
		return errors.WithStack(err)
	}

	return c.store.Save(storeKey, bytes.NewReader(data))
}

// Load - загружает фильтр из DataStore.
func (c *Component) Load() error {
	exists, err := c.store.IsExist(storeKey)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("no bloom filter data found")
	}

	dr, err := c.store.DataReader(storeKey)
	if err != nil {
		return err
	}
	defer func() {
		_ = dr.Close()
	}()

	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := io.ReadAll(dr)
	if err != nil {
		return err
	}

	err = c.impl.GobDecode(data)
	if err != nil {
		return errors.WithStack(err)
	}

	return c.store.Save(storeKey, bytes.NewReader(data))
}
