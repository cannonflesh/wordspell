package langdetect

import (
	"unicode"
)

const (
	numLangCode     = "num"
	enLangCode      = "en"
	ruLangCode      = "ru"
	unknownLangCode = "unknown"
)

type Component struct{}

func New() *Component {
	return &Component{}
}

func (c *Component) LangByWord(w string) string {
	runeWord := []rune(w)
	switch {
	case isNumber(runeWord):
		return numLangCode
	case isRus(runeWord):
		return ruLangCode
	case isEn(runeWord):
		return enLangCode
	}

	return unknownLangCode
}

func isNumber(rw []rune) bool {
	var pointFound bool
	for _, r := range rw {
		if r == '.' || r == ',' {
			if !pointFound {
				pointFound = true

				continue
			}

			return false
		}

		if !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func isRus(rw []rune) bool {
	var (
		legal   uint
		illegal uint
	)

	for _, r := range rw {
		if unicode.Is(unicode.Cyrillic, r) || r == '-' {
			legal++
		} else {
			illegal++
		}
	}

	return legal > illegal && illegal <= 2
}

func isEn(rw []rune) bool {
	var (
		legal   uint
		illegal uint
	)

	for _, r := range rw {
		if r <= unicode.MaxASCII && (unicode.IsLetter(r) || r == '-' || r == '`' || r == '\u0027') {
			legal++
		} else {
			illegal++
		}
	}

	return legal > illegal && illegal <= 2
}
