package dimensions

import (
	"regexp"
	"strings"

	"github.com/cannonflesh/wordspell/domain"
	"github.com/cannonflesh/wordspell/processors"
)

type Processor struct {
	requestRe   *regexp.Regexp
	separatorRe *regexp.Regexp
	tailRe      *regexp.Regexp
}

func New() *Processor {
	return &Processor{
		requestRe:   regexp.MustCompile(`(?:^|\s)(?:[\d.,]+\s?[xXхХ*/]\s?)+[\d.,]+(?:\s?(?:мм|см|дм|м|км|дюйм|mm|cm|m|km|in|ft))?`),
		separatorRe: regexp.MustCompile(`\s?[xXхХ*/]\s?`),
		tailRe:      regexp.MustCompile(`\s?(?:мм|см|дм|м|км|дюйм|mm|cm|m|km|in|ft)`),
	}
}

func (p *Processor) Process(words []string) []string {
	req := strings.Join(words, domain.SpaceSeparator)

	return strings.Fields(p.processStep(req))
}

func (p *Processor) processStep(req string) string {
	return p.requestRe.ReplaceAllStringFunc(req, func(in string) string {
		res := p.separatorRe.ReplaceAllString(in, "*")

		res = p.tailRe.ReplaceAllStringFunc(res, func(in string) string {
			return domain.SpaceSeparator + strings.TrimLeft(in, domain.SpaceSeparator)
		})

		prefix, body, suffix := processors.SplitChunk(res)

		return prefix + body + suffix
	})
}
