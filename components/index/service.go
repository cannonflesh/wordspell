package index

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cannonflesh/wordspell/domain"
	"github.com/cannonflesh/wordspell/options"
)

const (
	numLangCode     = "num"
	unknownLangCode = "unknown"
	numWeight       = 1000
)

type langDetector interface {
	LangByWord(word string) string
}

type Component struct {
	logger *logrus.Entry
	langs  langDetector
	mu     sync.RWMutex

	index map[string]map[string]uint32

	opt *options.Options
}

func New(opt *options.Options, langs langDetector, lgr *logrus.Entry) *Component {
	if len(opt.Langs) == 0 {
		opt.Langs = []string{"en", "ru"}
	}

	res := &Component{
		logger: lgr.WithField(domain.CategoryFieldName, "component.speller_index"),
		langs:  langs,

		index: make(map[string]map[string]uint32),

		opt: opt,
	}

	err := res.load(opt.Langs...)
	if err != nil {
		res.logger.WithError(err).Warn("loading index data from file")
	}

	return res
}

func (s *Component) Save(langs ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(langs) == 0 {
		langs = []string{"en", "ru"}
	}

	for _, lang := range langs {
		err := s.saveLangIndex(lang)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Component) saveLangIndex(lang string) error {
	var (
		idx   map[string]uint32
		found bool
		wfh   *os.File
		err   error
	)

	if idx, found = s.index[lang]; !found {
		return errors.New("no index for the such language: " + lang)
	}

	if wfh, err = s.indexFileWriteHandler(lang); err != nil {
		return err
	}
	defer func() {
		_ = wfh.Close()
	}()

	sorted := make([]*wordFrequency, 0, len(idx))
	for k, v := range idx {
		sorted = append(sorted, &wordFrequency{
			word:      k,
			frequency: v,
		})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].frequency > sorted[j].frequency
	})

	for _, line := range sorted {
		_, err = wfh.Write(line.toLine())
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (s *Component) load(langs ...string) error {
	for _, lang := range langs {
		err := s.parseFile(lang)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Component) SetLangIndex(lang string, idx map[string]uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.index[lang] = idx
}

func (s *Component) parseFile(lang string) error {
	path := filepath.Join(s.opt.DataDir, lang+".dat")
	fh, err := os.Open(path)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		_ = fh.Close()
	}()

	lineScan := bufio.NewScanner(fh)
	lineScan.Split(bufio.ScanLines)

	var idx map[string]uint32
	if idx = s.index[lang]; idx == nil {
		idx = make(map[string]uint32)
	}

	for lineScan.Scan() {
		word, frequency, err := s.parseFields(lineScan.Bytes())
		if err != nil {
			continue
		}

		idx[word] = idx[word] + frequency
	}

	if len(idx) > 0 {
		s.index[lang] = idx
	}

	return nil
}

func (s *Component) parseFields(in []byte) (string, uint32, error) {
	wordScan := bufio.NewScanner(bytes.NewBuffer(in))
	wordScan.Split(bufio.ScanWords)

	var (
		idx       int
		word      string
		frequency uint64
		err       error
	)
	for wordScan.Scan() {
		switch idx {
		case 0:
			word = wordScan.Text()
		case 1:
			frequency, err = strconv.ParseUint(wordScan.Text(), 10, 32)
			if err != nil {
				return "", 0, errors.WithStack(err)
			}
		}

		idx++
	}

	if word == "" {
		return word, uint32(frequency), errors.New("no word")
	}

	if frequency == 0 {
		return word, uint32(frequency), errors.New("no frequency")
	}

	return word, uint32(frequency), nil
}

func (s *Component) indexFileWriteHandler(lang string) (*os.File, error) {
	path := filepath.Join(s.opt.DataDir, lang+".dat")
	fh, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return fh, nil
}

func (s *Component) Weight(w string) uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lang := s.langs.LangByWord(w)
	if lang == unknownLangCode {
		s.logger.Warn("getting weight: language not detected")

		return 0
	}

	if lang == numLangCode {
		return numWeight
	}

	var (
		idx   map[string]uint32
		found bool
	)
	if idx, found = s.index[lang]; !found {
		s.logger.Error("getting weight: no index for such language: " + lang)

		return 0
	}

	return idx[w]
}

func (s *Component) DeletesEstimated(lang string) (uint, error) {
	if _, found := s.index[lang]; !found {
		return 0, errors.New("no index for the lang: " + lang)
	}

	var res uint
	for w := range s.index[lang] {
		wrl := runeLen(w)
		if wrl < 2 {
			continue
		}
		if wrl == 2 {
			res += 3
		}
		res += wrl*wrl + 1
	}

	return res, nil
}

func (s *Component) Words(lang string) (<-chan string, error) {
	if _, found := s.index[lang]; !found {
		return nil, errors.New("no index for the lang: " + lang)
	}

	res := make(chan string)

	go func() {
		for w := range s.index[lang] {
			res <- w
		}

		close(res)
	}()

	return res, nil
}

func runeLen(w string) uint {
	return uint(len([]rune(w)))
}
