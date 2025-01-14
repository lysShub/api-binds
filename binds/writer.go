package binds

import (
	"bytes"
	"strconv"
)

type writer struct {
	b bytes.Buffer
}

func (b *writer) Bytes() []byte { return b.b.Bytes() }
func (b *writer) Reset()        { b.b.Reset() }
func (b *writer) Str(s string) *writer {
	b.b.WriteString(s)
	return b
}
func (b *writer) Int(i int) *writer {
	return b.Str(strconv.Itoa(i))
}
func (b *writer) Char(c byte) *writer {
	b.b.WriteByte(c)
	return b
}
func (b *writer) NL() *writer { return b.Char('\n') }
func (b *writer) SP() *writer { return b.Char(' ') }
func (b *writer) EQ() *writer { return b.Char('=') }
func (b *writer) PT() *writer { return b.Char('.') }
func (b *writer) DS() *writer { return b.Str("//") }
func (b *writer) QM() *writer { return b.Char('?') }
func (b *writer) US() *writer { return b.Char('_') }

func (b *writer) Unchar() *writer {
	b.b.Truncate(b.b.Len() - 1)
	return b
}
