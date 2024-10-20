package main

import (
	"bytes"
	"html"
	"net/url"
	"path"
	"strings"

	"github.com/barnslig/bilbo/goldmark/wikitag"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"go.abhg.dev/goldmark/anchor"
)

type wikiTagResolver struct {
	pageExists func(path string) bool
}

func (r *wikiTagResolver) Resolve(w util.BufWriter, n *wikitag.Node) (err error) {
	c := strings.SplitN(string(n.Content), "|", 2)

	label := html.EscapeString(c[0])

	target := c[0]
	if len(c) == 2 {
		target = c[1]
	}

	link, err := url.Parse(target)
	if err != nil {
		return
	}

	classes := "external"
	if !link.IsAbs() {
		classes = "internal"
		link.Path = normalizePageLink(link.Path, true)

		if !r.pageExists(link.Path) {
			classes = "internal absent"
		}
	}

	w.WriteString("<a href=\"")
	w.WriteString(link.String())
	w.WriteString("\" class=\"")
	w.WriteString(classes)
	w.WriteString("\">")
	w.WriteString(label)
	w.WriteString("</a>")

	return nil
}

func extractTitle(doc ast.Node, source []byte) (title string, err error) {
	err = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		node, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		if node.Level == 1 {
			title = string(node.Text(source))
			return ast.WalkStop, nil
		}

		return ast.WalkContinue, nil
	})

	return
}

func (b *Bilbo) RenderPage(page *Page, commit plumbing.Hash) (err error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&anchor.Extender{
				Texter: anchor.Text("#"),
			},
			wikitag.NewWikiTag(wikitag.WikiTagConfig{
				Resolver: &wikiTagResolver{
					pageExists: func(path string) bool {
						_, err := b.getPageAtCommit(path, false, commit)
						return err != nil
					},
				},
			}),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	// Render the page according to its extension
	ext := path.Ext(page.Filepath)
	switch ext {
	case ".md", ".markdown":
		doc := md.Parser().Parse(text.NewReader(page.Source))

		var buf bytes.Buffer
		err = md.Renderer().Render(&buf, page.Source, doc)
		if err != nil {
			return
		}

		page.Title, _ = extractTitle(doc, page.Source)
		page.Rendered = buf.Bytes()
	default:
		page.Rendered = page.Source
	}

	return
}
