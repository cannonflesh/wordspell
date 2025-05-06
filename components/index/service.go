package index

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"sync"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cannonflesh/wordspell/domain"
	"github.com/cannonflesh/wordspell/options"
)

type langDetector interface {
	LangByWord(word string) string
	ParseWordPair(pair []string) (string, string, string)
}

// DataStore описывает методы чтения-записи так,
// чтобы поменьше хранить в памяти приложения (ридеры позволяют читать данные порциями).
// Данных ожидается много, но в конце концов нам придется все их загрузить в индексы.
//
// Если мы будем читать данные индексов и генерить bitmap для bloom-фильтра
// одновременно со штатной работой wordspell (когда он хранит актуальные индексы и фильтр),
// мы потратим удвоенную память только на обновление индексов и фильтра.
//
// Поэтому предлагаю запускать индексатор раз в неделю в отдельном поде по крону,
// а когда новые индексы будут построены, "крутануть" деплойменты приложения,
// использующего wordspell, тоже по крону.
// Тогда спелл-чекер на старте заполнит свои пустые на тот момент индексы из DataStore
// и построит себе новый bloom-фильтр. Битмап bloom-фильтра под одними и теми же настройками
// зависит только от индексов, строится относительно быстро, и его можно создавать "на лету".
//
// Таким образом памяти будет использовано ровно столько, сколько нужно.
type DataStore interface {
	DataReader(key string) (io.ReadCloser, error)
	IsExist(key string) (bool, error)
	Save(key string, content io.Reader) error
}

type Service struct {
	logger *logrus.Entry
	langs  langDetector
	mu     sync.RWMutex

	store DataStore
	index wordCollection

	opt *options.Options
}

func NewService(
	opt *options.Options,
	langs langDetector,
	store DataStore,
	lgr *logrus.Entry,
) (*Service, error) {
	if len(opt.Langs) == 0 {
		opt.Langs = []string{enLangCode, ruLangCode}
	}

	res := &Service{
		logger: lgr.WithField(domain.CategoryFieldName, "component.speller_index_service"),
		langs:  langs,

		store: store,
		index: make(wordCollection),

		opt: opt,
	}

	if err := res.load(); err != nil {
		return nil, err
	}

	return res, nil
}

// Weight рабочий метод индекса, используемый спеллером.
func (s *Service) Weight(w string) uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lang := s.langs.LangByWord(w)
	if lang == unknownLangCode {
		s.logger.Debug("getting weight: language not detected")

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

// SetLangIndex записывает новые данные в индекс переданного языка.
// Используется для загрузки индекса после запуска приложения.
func (s *Service) SetLangIndex(lang string, idx map[word]frequency) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.index[lang] = idx
}

// DeletesEstimated - используется для расчета bitmap bloom-фильтра.
func (s *Service) DeletesEstimated() (uint, error) {
	var res uint
	for lang := range s.index {
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
	}

	return res, nil
}

// Words - используется для расчета bitmap bloom-фильтра.
func (s *Service) Words() (<-chan string, error) {
	res := make(chan string)

	go func() {
		for lang := range s.index {
			for w := range s.index[lang] {
				res <- w
			}
		}

		close(res)
	}()

	return res, nil
}

// runeLen отдает длину строки в рунах.
// Реально len(string) отдает длину строки в байтах.
// См. подтверждение в тестах.
func runeLen(w string) uint {
	return uint(len([]rune(w)))
}

// load загружает индексы из store в конструкторе индекса.
func (s *Service) load() error {
	for _, lang := range s.opt.Langs {
		err := s.parseData(lang)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) parseData(l string) error {
	dh, err := s.store.DataReader(langCodeIndexKey(l))
	if err != nil {
		return err
	}

	defer func() {
		_ = dh.Close()
	}()

	lineScan := bufio.NewScanner(dh)
	lineScan.Split(bufio.ScanLines)

	var idx map[string]uint32
	if idx = s.index[l]; idx == nil {
		idx = make(map[string]uint32)
	}

	for lineScan.Scan() {
		w, f, err := s.parseFields(lineScan.Bytes())
		if err != nil {
			s.logger.WithError(err).Warn("scanning index data line")

			continue
		}

		idx[w] = idx[w] + f
	}

	if len(idx) > 0 {
		s.index[l] = idx
	}

	return nil
}

func (s *Service) parseFields(in []byte) (word, frequency, error) {
	wordScan := bufio.NewScanner(bytes.NewBuffer(in))
	wordScan.Split(scanTabDelimited)

	var (
		idx int
		w   string
		f   uint64
		err error
	)
	for wordScan.Scan() {
		switch idx {
		case 0:
			w = wordScan.Text()
		case 1:
			f, err = strconv.ParseUint(wordScan.Text(), 10, 32)
			if err != nil {
				return "", 0, errors.WithStack(err)
			}
		}

		idx++
	}

	if w == "" {
		return w, uint32(f), errors.New("no word")
	}

	if f == 0 {
		return w, uint32(f), errors.New("no frequency")
	}

	return w, uint32(f), nil
}

// scanTabDelimited сделана по образу bufio.ScanWord, но не удаляет пробелы с начала строки
// (у нас их быть не может), и разбивает строки по табулятору \t.
func scanTabDelimited(data []byte, atEOF bool) (advance int, token []byte, err error) {
	start := 0
	// Scan until tab, marking end of word.
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if r == '\t' {
			return i + width, data[start:i], nil
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// Request more data.
	return start, nil, nil
}
