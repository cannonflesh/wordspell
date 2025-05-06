package langdetect

import (
	"strings"
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

func (c *Component) ParseWordPair(pair []string) (string, string, string) {
	if len(pair) == 0 {
		return "", "", unknownLangCode
	}

	first := strings.ToLower(pair[0])
	if first == "" {
		return "", "", unknownLangCode
	}

	if !badRuWord(first) {
		if len(pair) == 1 {
			return first, "", ruLangCode
		}

		second := strings.ToLower(pair[1])

		if badRuWord(second) {
			return first, "", ruLangCode
		}

		return first, second, ruLangCode
	}

	if !badEnWord(first) {
		if len(pair) == 1 {
			return first, "", enLangCode
		}

		second := strings.ToLower(pair[1])

		if badEnWord(second) {
			return first, "", enLangCode
		}

		return first, second, enLangCode
	}

	return "", "", unknownLangCode
}

func badRuWord(w string) bool {
	if w == "" {
		return true
	}

	wr := []rune(w)
	for _, r := range wr {
		if !unicode.Is(unicode.Cyrillic, r) && r != '-' {
			return true
		}
	}

	return false
}

func badEnWord(w string) bool {
	if w == "" {
		return true
	}

	wr := []rune(w)
	for _, r := range wr {
		if r > unicode.MaxASCII || (!unicode.IsLetter(r) && r != '-' && r != '`' && r != '\u0027') {
			return true
		}
	}

	return false
}
