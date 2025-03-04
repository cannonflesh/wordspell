package index

import (
	"bytes"
	"strconv"
)

type wordFrequency struct {
	word      string
	frequency uint32
}

func (wf *wordFrequency) toLine() []byte {
	var buf bytes.Buffer

	buf.WriteString(wf.word)
	buf.WriteString("\t")
	buf.WriteString(strconv.FormatUint(uint64(wf.frequency), 10))
	buf.WriteString("\n")

	return buf.Bytes()
}
