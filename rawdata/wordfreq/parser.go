package wordfreq

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cannonflesh/wordspell/domain"
)

const (
	wikiEn = `eng-simple_wikipedia_2021_300K-words.txt`
	wikiRu = `rus_wikipedia_2021_300K-words.txt`
	newsEn = `eng_news_2024_300K-words.txt`
	newsRu = `rus_news_2024_300K-words.txt`
)

type Parser struct {
	logger *logrus.Entry
}

func New(l *logrus.Entry) *Parser {
	return &Parser{
		logger: l.WithField(domain.CategoryFieldName, "file_parser.word_frequency"),
	}
}

func (p *Parser) BuildLangIndex(lang string) (map[string]uint32, error) {
	res := make(map[string]uint32)
	dirPath := fileDir()

	var wikiPath, newsPath string
	if lang == domain.EnLangCode {
		wikiPath = filepath.Join(dirPath, wikiEn)
		newsPath = filepath.Join(dirPath, newsEn)
	} else if lang == domain.RuLangCode {
		wikiPath = filepath.Join(dirPath, wikiRu)
		newsPath = filepath.Join(dirPath, newsRu)
	} else {
		return nil, errors.New("incorrect language: " + lang)
	}

	start := time.Now()

	wikiTotal, wikiMissed, err := p.parseFile(res, wikiPath, lang)
	if err != nil {
		return nil, err
	}

	newsTotal, newsMissed, err := p.parseFile(res, newsPath, lang)
	if err != nil {
		return nil, err
	}

	p.logger.Infof(
		"%s index loaded in %s, %d words loaded, %d words missed",
		lang, time.Since(start), wikiTotal+newsTotal, wikiMissed+newsMissed,
	)

	return res, nil
}

func (p *Parser) parseFile(index map[string]uint32, path string, lang string) (int, int, error) {
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

	total := 0
	missed := 0
	for lineScan.Scan() {
		word, frequency, err := p.parseFields(lineScan.Bytes(), lang)
		if err != nil {
			missed++

			continue
		}

		currentFreq := index[word]
		index[word] = currentFreq + frequency

		if currentFreq == 0 {
			total++
		}
	}

	return total, missed, nil
}

func (p *Parser) parseFields(in []byte, lang string) (string, uint32, error) {
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
		case 1:
			word = strings.ToLower(wordScan.Text())
		case 2:
			frequency, err = strconv.ParseUint(wordScan.Text(), 10, 32)
			if err != nil {
				return "", 0, errors.WithStack(err)
			}
		}

		idx++
	}

	if badWord(word, lang) || frequency == 0 {
		return word, uint32(frequency), errors.New("bad word or no frequency")
	}

	return word, uint32(frequency), nil
}

func badWord(w string, lang string) bool {
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
