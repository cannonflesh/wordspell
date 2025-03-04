package testdata

import (
	"bytes"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

func NewTestLogger() (*logrus.Entry, *Buffer) {
	buf := &Buffer{}
	l := logrus.New()
	l.SetOutput(buf)
	log := logrus.NewEntry(l)
	return log, buf
}

type Buffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}
func (b *Buffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}

func (b *Buffer) Lines() []string {
	res := strings.Split(b.String(), "\n")
	return res[:len(res)-1]
}

func (b *Buffer) Count() int {
	return len(b.Lines())
}
