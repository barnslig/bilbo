package main

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/barnslig/bilbo/goldmark/wikitag"
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

	if len(title) == 0 {
		err = errors.New("title not found")
	}

	return
}

type anchorAttributor struct{}

func (aa anchorAttributor) AnchorAttributes(headerInfo *anchor.HeaderInfo) map[string]string {
	return map[string]string{
		"class": "a-hl-anchor",
		"title": fmt.Sprintf("Set anchor to \"%s\"", headerInfo.ID),
	}
}

type MarkdownPageRendererOptions struct {
	pageExists func(path string) bool
}

type markdownPageRenderer struct {
	md goldmark.Markdown
}

func NewMarkdownPageRenderer(options MarkdownPageRendererOptions) PageRenderer {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&anchor.Extender{
				Texter:     anchor.Text("🔗"),
				Attributer: anchorAttributor{},
			},
			wikitag.NewWikiTag(wikitag.WikiTagConfig{
				Resolver: &wikiTagResolver{
					pageExists: options.pageExists,
				},
			}),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	return &markdownPageRenderer{
		md: md,
	}
}

func (r *markdownPageRenderer) Render(page *Page) (err error) {
	doc := r.md.Parser().Parse(text.NewReader(page.Source))

	var buf bytes.Buffer
	err = r.md.Renderer().Render(&buf, page.Source, doc)
	if err != nil {
		return
	}

	if title, err := extractTitle(doc, page.Source); err == nil {
		page.Title = title
	}

	page.Rendered = buf.Bytes()

	return
}
