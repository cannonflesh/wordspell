package processors

import (
	"strings"

	"github.com/cannonflesh/wordspell/domain"
)

func SplitChunk(chunk string) (string, string, string) {
	prefix := "@"
	suffix := ""
	if strings.HasSuffix(chunk, domain.SpaceSeparator) {
		suffix = domain.SpaceSeparator
	}
	if strings.HasPrefix(chunk, domain.SpaceSeparator) {
		prefix = domain.SpaceSeparator + prefix
	}

	return prefix,
		strings.Replace(strings.TrimSpace(chunk), domain.SpaceSeparator, domain.ComboSeparator, -1),
		suffix
}
