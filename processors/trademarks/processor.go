package trademarks

type finder interface {
	Find(hayStack []string) (string, []string)
}

type Processor struct {
	tradeMarks finder
}

func New(tmf finder) *Processor {
	return &Processor{tradeMarks: tmf}
}

func (p *Processor) Process(words []string) []string {
	if p.tradeMarks == nil {
		return words
	}

	var (
		head string
		tail []string
		res  []string
	)

	if len(words) == 0 {
		return words
	}

	for {
		head, tail = p.tradeMarks.Find(words)
		if head == "" {
			res = append(res, words[0])

			words = words[1:]
			if len(words) == 0 {
				return res
			}

			continue
		}

		break
	}

	res = append(res, head)
	res = append(res, p.Process(tail)...)

	return res
}
