package dimsuffix

import (
	"regexp"
	"strings"

	"github.com/cannonflesh/wordspell/domain"
	"github.com/cannonflesh/wordspell/processors"
)

type Processor struct {
	re *regexp.Regexp
}

func New() *Processor {
	return &Processor{
		re: regexp.MustCompile(`(?:^|\s)[2-5]\s?[dDдД](?:\s|$)`),
	}
}

func (p *Processor) Process(words []string) []string {
	req := strings.Join(words, domain.SpaceSeparator)
	res := p.processStep(p.processStep(req)) // для обработки шаблонов, идущих подряд

	return strings.Fields(res)
}

func (p *Processor) processStep(req string) string {
	return p.re.ReplaceAllStringFunc(req, func(in string) string {
		prefix, body, suffix := processors.SplitChunk(in)

		return prefix + string(body[0]) + "D" + suffix
	})
}
