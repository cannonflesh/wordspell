package units

import (
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/domain"
	"gitlab.sima-land.ru/dev-dep/dev/packages/go-wordspell/processors"
	"regexp"
	"strings"
)

type Processor struct {
	unitsRe        *regexp.Regexp
	unitsPrefixRe  *regexp.Regexp
	unitsHyphenRe  *regexp.Regexp
	unitsEqualSign *regexp.Regexp
	unitsTailRe    *regexp.Regexp
}

func New() *Processor {
	return &Processor{
		unitsRe:        regexp.MustCompile(`(?:^|\s)(?:(?:l|d|r)\s?=?)?\s?(?:(?:[\d.,]+\s?%?)\s?-\s?)*(?:[\d.,]+\s?%?)(?:\s?(?:мм|см|дм|м|км|д|дюйм|mm|cm|m|km|in|ft|кв мм|кв см|кв м|кв км|sq mm|sq cm|sq m|sq km|sq in|sq ft|мм2|см2|м2|км2|д2|дюйм2|mm2|cm2|m2|km2|in2|ft2|куб мм|куб см|куб м|куб км|куб д|куб дюйм|мм3|см3|м3|км3|д3|дюйм3|mm3|cm3|m3|km3|in3|ft3|мл|л|мг|г|кг|в|вт|ом|ком|рад|град|шт))?`),
		unitsPrefixRe:  regexp.MustCompile(`[ldr]\s`),
		unitsHyphenRe:  regexp.MustCompile(`\s?-\s?`),
		unitsEqualSign: regexp.MustCompile(`\s?=\s?`),
		unitsTailRe:    regexp.MustCompile(`\s?(?:мм|см|дм|м|км|дюйм|mm|cm|m|km|in|ft|кв мм|кв см|кв м|кв км|кв дюйм|sq mm|sq cm|sq m|sq km|sq in|sq ft|мм2|см2|м2|км2|дюйм2|mm2|cm2|m2|km2|in2|ft2|куб мм|куб см|куб м|куб км|куб дюйм|мм3|см3|м3|км3|дюйм3|mm3|cm3|m3|km3|in3|ft3|мл|л|мг|г|кг|в|вт|ом|ком|рад|град|шт)`),
	}
}

func (p *Processor) Process(words []string) []string {
	req := strings.Join(words, domain.SpaceSeparator)

	return strings.Fields(p.processStep(req))
}

func (p *Processor) processStep(req string) string {
	return p.unitsRe.ReplaceAllStringFunc(req, func(outIn string) string {
		res := p.unitsHyphenRe.ReplaceAllString(outIn, "-")
		res = p.unitsEqualSign.ReplaceAllString(res, "=")

		res = p.unitsPrefixRe.ReplaceAllStringFunc(res, func(in string) string {
			return strings.TrimSpace(in)
		})

		res = p.unitsTailRe.ReplaceAllStringFunc(res, func(in string) string {
			return domain.SpaceSeparator + strings.TrimLeft(in, domain.SpaceSeparator)
		})

		prefix, body, suffix := processors.SplitChunk(res)

		return prefix + strings.Replace(body, "#%", "%", -1) + suffix
	})
}
