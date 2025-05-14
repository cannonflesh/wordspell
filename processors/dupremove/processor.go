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

		checkWord := strings.ToLower(words[i])
		if checkWord == left || checkWord == right {
			continue
		}

		last = words[i]
		res = append(res, last)
	}

	return res
}

func rightChunk(word string) string {
	splitted := strings.Split(word, "-")
	return strings.ToLower(splitted[len(splitted)-1])
}

func leftChunk(word string) string {
	splitted := strings.Split(word, "-")
	return strings.ToLower(splitted[0])
}
