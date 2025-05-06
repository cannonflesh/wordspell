package domain

import "regexp"

const (
	RuLangCode      = "ru"
	EnLangCode      = "en"
	NumLangCode     = "num"
	UnknownLangCode = "unknown"
)

const CategoryFieldName = "category"

const (
	SpaceSeparator = " "
	ComboSeparator = "#"
	ComboPrefix    = "@"
)

var CleanTextRE = regexp.MustCompile("\\s-\\s|[^0-9a-zA-Zа-яА-ЯёЁ\\s-.,+=`'*%]+")
var CleanIndexRE = regexp.MustCompile("\\s-\\s|[^a-zA-Zа-яА-ЯёЁ\\s-`']+")
