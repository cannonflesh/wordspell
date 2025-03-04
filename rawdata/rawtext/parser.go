package rawtext

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cannonflesh/wordspell/domain"
)

type Parser struct {
	logger *logrus.Entry
}

func New(l *logrus.Entry) *Parser {
	return &Parser{
		logger: l.WithField(domain.CategoryFieldName, "file_parser.raw_text"),
	}
}

func (p *Parser) BuildLangIndex(lang string) (map[string]uint32, error) {
	res := make(map[string]uint32)

	start := time.Now()

	total, missed, err := p.parseFile(res, lang)
	if err != nil {
		return nil, err
	}

	p.logger.Infof(
		"%s index loaded in %s, %d words loaded, %d words missed",
		lang, time.Since(start), total, missed,
	)

	return res, nil
}

func (p *Parser) parseFile(idx map[string]uint32, lang string) (int, int, error) {
	path := filepath.Join(fileDir(), lang+`.txt`)

	fh, err := os.Open(path)
	if err != nil {
		if err != nil {
			return 0, 0, errors.WithStack(err)
		}
	}
	defer func() {
		_ = fh.Close()
	}()

	lineScan := bufio.NewScanner(fh)
	lineScan.Split(bufio.ScanLines)

	var total, missed int
	for lineScan.Scan() {
		lineTotal, lineMissed := p.parseWords(idx, lineScan.Bytes(), lang)

		total += lineTotal
		missed += lineMissed
	}

	return total, missed, nil
}

func (p *Parser) parseWords(idx map[string]uint32, in []byte, lang string) (int, int) {
	wordScan := bufio.NewScanner(bytes.NewBuffer(in))
	wordScan.Split(bufio.ScanWords)

	var total, misses int

	var word string
	for wordScan.Scan() {
		word = strings.ToLower(wordScan.Text())
		if badWord(word, lang) {
			misses++

			continue
		}

		currentFreq := idx[word]

		idx[word] = currentFreq + 1
		if currentFreq == 0 {
			total++
		}
	}

	return total, misses
}

func badWord(w string, lang string) bool {
	if w == `-` {
		return true
	}

	runeStr := []rune(w)

	if lang == domain.RuLangCode {
		return badRuWord(runeStr)
	} else if lang == domain.EnLangCode {
		return badEnWord(runeStr)
	}

	return true
}

func badRuWord(w []rune) bool {
	for _, r := range w {
		if !unicode.Is(unicode.Cyrillic, r) && r != '-' {
			return true
		}
	}

	return false
}

func badEnWord(w []rune) bool {
	for _, r := range w {
		if r > unicode.MaxASCII || (!unicode.IsLetter(r) && r != '-') {
			return true
		}
	}

	return false
}

func fileDir() string {
	_, fp, _, _ := runtime.Caller(0)

	return filepath.Dir(fp)
}
