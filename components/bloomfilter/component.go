package bloomfilter

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/cannonflesh/wordspell/domain"
	"github.com/cannonflesh/wordspell/options"
)

const (
	dataFileName             = "bloom.dat"
	defaultLength            = 100000
	defaultFalsePositiveRate = 0.005
)

type Component struct {
	opt    *options.Options
	impl   *bloom.BloomFilter
	logger *logrus.Entry
}

func New(opt *options.Options, logger *logrus.Entry) *Component {
	res := &Component{
		opt:    opt,
		logger: logger.WithField(domain.CategoryFieldName, "components.bloom_filter"),
	}

	err := res.load()
	if err != nil {
		res.logger.WithError(err).Warn("loading bloom filter data from file")
	}

	return res
}

func (c *Component) Reset(size uint) {
	c.impl = bloom.NewWithEstimates(size, defaultFalsePositiveRate)
}

func (c *Component) Add(words ...string) {
	for _, w := range words {
		_ = c.impl.Add([]byte(w))
	}
}

func (c *Component) Test(w string) bool {
	return c.impl.TestString(w)
}

func (c *Component) Save() error {
	path := filepath.Join(c.opt.DataDir, dataFileName)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = c.impl.WriteTo(f)
	if err != nil {
		return errors.WithStack(err)
	}

	err = f.Sync()

	return errors.WithStack(err)
}

func (c *Component) load() error {
	path := filepath.Join(c.opt.DataDir, dataFileName)
	f, err := os.Open(path)
	if err != nil {
		c.impl = bloom.NewWithEstimates(defaultLength, defaultFalsePositiveRate)

		return errors.WithStack(err)
	}

	if c.impl == nil {
		c.impl = &bloom.BloomFilter{}
	}
	r := bufio.NewReader(f)
	btr, err := c.impl.ReadFrom(r)
	c.logger.Infof("%d bytes read from the bloom.dat file", btr)

	return errors.WithStack(err)
}
