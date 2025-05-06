package index

import (
	"bytes"
	"strconv"
)

const numWeight = 1000

// wordFrequency применяется для (де-)сериализации мндексов.
type wordFrequency struct {
	word      string
	frequency uint32
}

func (wf *wordFrequency) toLine() []byte {
	var buf bytes.Buffer

	buf.WriteString(wf.word)
	buf.WriteString("\t")
	buf.WriteString(strconv.FormatUint(uint64(wf.frequency), 10))
	buf.WriteString("\n")

	return buf.Bytes()
}

type (
	langCode       = string
	word           = string
	frequency      = uint32
	wordCollection map[langCode]map[word]frequency
	data           struct {
		words  wordCollection
		dwords wordCollection
	}
)

const (
	numLangCode     langCode = "num"
	ruLangCode      langCode = "ru"
	enLangCode      langCode = "en"
	unknownLangCode langCode = "unknown"
)

func newData() *data {
	return &data{
		words: wordCollection{
			enLangCode: make(map[word]frequency),
			ruLangCode: make(map[word]frequency),
		},
		dwords: wordCollection{
			enLangCode: make(map[word]frequency),
			ruLangCode: make(map[word]frequency),
		},
	}
}

func (d *data) merge(add *data) {
	d.words.merge(add.words)
	d.dwords.merge(add.dwords)
}

func (wc wordCollection) merge(add wordCollection) {
	for k, v := range add {
		if wc[k] == nil {
			wc[k] = make(map[word]frequency, len(v))
		}
		for kk, vv := range v {
			wc[k][kk] = wc[k][kk] + vv
		}
	}
}
