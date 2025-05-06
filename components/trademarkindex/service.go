package trademarkindex

import (
	"bufio"
	"io"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cannonflesh/wordspell/domain"
)

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

const storeKey = "trademark.index"

type Service struct {
	store  DataStore
	index  tradeMarkIndex
	logger *logrus.Entry

	mu sync.RWMutex
}

func NewService(store DataStore, lgr *logrus.Entry) (*Service, error) {
	res := &Service{
		store:  store,
		logger: lgr.WithField(domain.CategoryFieldName, "component.trademarks_index_service"),
		mu:     sync.RWMutex{},
	}

	idx, err := res.loadIndex()
	if err != nil {
		return nil, err
	}

	res.index = idx

	return res, nil
}

// Find отыскивает трейдмарки по соответствующему словарю, форматирует их в одиночные слова
// в формате @Apple#inc#Ltd, чтобы следующим ходом, убрав лишние символы, пометить их как элементы,
// не подлежащие дальнейшей обработке.
func (s *Service) Find(hayStack []string) (string, []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(hayStack) == 0 {
		return "", nil
	}

	var (
		tails [][]string
		found bool
	)

	head := hayStack[0]
	if tails, found = s.index[head]; !found {
		return "", hayStack
	}

	var (
		oneWordAllowed bool
		maxTailLen     int
		hayTail        = hayStack[1:]
		tradeMark      = head
	)
	for _, tail := range tails {
		tailLen := len(tail)
		if tailLen == 0 {
			oneWordAllowed = true
		}
		if tailLen > len(hayTail) || tailLen < maxTailLen {
			continue
		}

		check := strings.Join(hayTail[:tailLen], domain.ComboSeparator)
		if check == strings.Join(tail, domain.ComboSeparator) {
			tradeMark = strings.Join(hayStack[:tailLen+1], domain.ComboSeparator)

			if tailLen > maxTailLen {
				maxTailLen = tailLen
			}
		}
	}

	if maxTailLen == 0 && !oneWordAllowed {
		return "", hayStack
	}

	if tradeMark != "" {
		tradeMark = domain.ComboPrefix + tradeMark
	}

	return tradeMark, hayTail[maxTailLen:]
}

func (s *Service) loadIndex() (tradeMarkIndex, error) {
	res := make(tradeMarkIndex)

	if ok, err := s.store.IsExist(storeKey); err != nil || !ok {
		if err == nil {
			err = errors.New("no trademark index found")
		}

		return nil, err
	}

	data, err := s.store.DataReader(storeKey)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = data.Close()
	}()

	scan := bufio.NewScanner(data)
	scan.Split(bufio.ScanLines)

	for scan.Scan() {
		line := scan.Text()
		parseName(line, res)
	}

	return res, nil
}

func parseName(line string, idx tradeMarkIndex) {
	nameWords := strings.Fields(line)

	if len(nameWords) == 0 {
		return
	}

	head := nameWords[0]

	tail := make(tradeMarkTail, 0, 0)
	if len(nameWords) > 1 {
		tail = nameWords[1:]
	}

	idx[head] = append(idx[head], tail)
}
