package wikitag

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type wikiTagRenderer struct {
	config WikiTagConfig
}

func NewWikiTagRenderer(config WikiTagConfig) renderer.NodeRenderer {
	r := &wikiTagRenderer{config: config}
	return r
}

func (r *wikiTagRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(Kind, r.Render)
}

func (r *wikiTagRenderer) Render(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	node, ok := n.(*Node)
	if !ok {
		return ast.WalkStop, fmt.Errorf("unexpected node %T, expected *goldmarkwikitag.Node", n)
	}

	if entering {
		err := r.config.Resolver.Resolve(w, node)
		if err != nil {
			return ast.WalkStop, err
		}
	}

	return ast.WalkContinue, nil
}
