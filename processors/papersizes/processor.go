package papersizes

import (
	"regexp"
	"strings"

	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/domain"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/processors"
)

type Processor struct {
	paperSizes   map[string]string
	paperSizesRe *regexp.Regexp
}

func New() *Processor {
	return &Processor{
		paperSizes: map[string]string{
			"а": "A",
			"А": "A",
			"a": "A",
			"A": "A",
			"b": "B",
			"B": "B",
			"В": "B",
		},
		paperSizesRe: regexp.MustCompile(`(?:^|\s)[aAаАbBВ]\s?[0-6]`),
	}
}

func (p *Processor) Process(words []string) []string {
	req := strings.Join(words, domain.SpaceSeparator)
	res := p.processStep(req)

	return strings.Fields(res)
}

func (p *Processor) processStep(req string) string {
	return p.paperSizesRe.ReplaceAllStringFunc(req, func(paperIn string) string {
		pre, chunk, suf := processors.SplitChunk(paperIn)
		runePair := []string{
			p.paperSizes[string([]rune(chunk)[0])],
			string([]rune(chunk)[len([]rune(chunk))-1]),
		}

		return pre + strings.ToUpper(strings.Join(runePair, "")) + suf
	})
}
