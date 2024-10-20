package wikitag

import "github.com/yuin/goldmark/ast"

var Kind = ast.NewNodeKind("WikiTag")

type Node struct {
	ast.BaseInline

	Content []byte
}

func (n *Node) Kind() ast.NodeKind {
	return Kind
}

func (n *Node) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, map[string]string{
		"Content": string(n.Content),
	}, nil)
}
