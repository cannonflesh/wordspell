package domain

import "strings"

type Digest []DigestElement

func NewEmptyDigest() Digest {
	return make(Digest, 0)
}

func ParseDigest(words []string) Digest {
	var res Digest

	for _, w := range words {
		if strings.HasPrefix(w, ComboPrefix) {
			elText := strings.Replace(w, ComboPrefix, "", 1)
			res = append(res, NewDigestReady(strings.Replace(elText, ComboSeparator, SpaceSeparator, -1)))

			continue
		}

		res = append(res, NewDigestRaw(w))
	}

	return res
}

func (d Digest) Add(els ...DigestElement) Digest {
	return append(d, els...)
}

type DigestElement interface {
	isDigestElement()
	String() string
}

type DigestRaw string

func NewDigestRaw(s string) DigestRaw {
	return DigestRaw(s)
}

func (draw DigestRaw) Merge(right DigestRaw) DigestRaw {
	return draw + right
}

func (draw DigestRaw) String() string {
	return string(draw)
}

func (draw DigestRaw) isDigestElement() {}

type DigestReady string

func NewDigestReady(s string) DigestReady {
	return DigestReady(s)
}

func (dr DigestReady) String() string {
	return string(dr)
}

func (dr DigestReady) isDigestElement() {}
