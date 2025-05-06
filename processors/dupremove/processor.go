package dupremove

import "strings"

type Processor struct{}

func New() *Processor {
	return new(Processor)
}

func (p *Processor) Process(words []string) []string {
	res := make([]string, 0, len(words))
	lastIdx := len(words) - 1
	last := ""

	for i := range words {
		var left, right string

		if last != "" {
			left = rightChunk(last)
		}

		if i < lastIdx {
			right = leftChunk(words[i+1])
		}

		if words[i] == left || words[i] == right {
			continue
		}

		last = words[i]
		res = append(res, last)
	}

	return res
}

func rightChunk(word string) string {
	splitted := strings.Split(word, "-")
	return splitted[len(splitted)-1]
}

func leftChunk(word string) string {
	splitted := strings.Split(word, "-")
	return splitted[0]
}
