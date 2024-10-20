package wikitag

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type WikiTagResolver interface {
	Resolve(w util.BufWriter, n *Node) error
}

type defaultWikiTagResolver struct{}

func (r *defaultWikiTagResolver) Resolve(w util.BufWriter, n *Node) error {
	w.WriteString("<code>")
	w.Write(n.Content)
	w.WriteString("</code>")

	return nil
}

var DefaultWikiTagResolver = defaultWikiTagResolver{}

type WikiTagConfig struct {
	Resolver WikiTagResolver
}

type wikiTag struct {
	config WikiTagConfig
}

func NewWikiTag(config WikiTagConfig) goldmark.Extender {
	return &wikiTag{config: config}
}

func (e *wikiTag) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewWikiTagParser(), 199),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewWikiTagRenderer(e.config), 199),
	))
}
