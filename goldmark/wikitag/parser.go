package wikitag

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var (
	_open  = []byte("[[")
	_close = []byte("]]")
)

type wikiTagParser struct{}

func NewWikiTagParser() parser.InlineParser {
	return &wikiTagParser{}
}

func (p *wikiTagParser) Trigger() []byte {
	return []byte{'['}
}

func (p *wikiTagParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, seg := block.PeekLine()

	stop := bytes.Index(line, _close)
	if stop < 0 {
		return nil // must close on the same line
	}

	if bytes.HasPrefix(line, _open) {
		seg = text.NewSegment(seg.Start+len(_open), seg.Start+stop)
	}

	block.Advance(stop + len(_close))

	n := &Node{Content: block.Value(seg)}
	return n
}
