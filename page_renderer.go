package main

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"gopkg.in/russross/blackfriday.v2"
	"io"
	"net/url"
	"path"
	"regexp"
	"strings"
)

var (
	headlineRegex = regexp.MustCompile(`^(?i)h[0-9]$`)
	linkRegex     = regexp.MustCompile(`\[\[(.+)\|(.+)\]\]|\[\[(.+)\]\]`)
)

func htmlNodeExtractString(n *html.Node) (str string) {
	var (
		b strings.Builder
		f func(*html.Node)
	)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return b.String()
}

func htmlNodeGetAttributeByKey(n *html.Node, key string) (val string, err error) {
	for _, attr := range n.Attr {
		if attr.Key == key {
			val = attr.Val
			return
		}
	}

	err = fmt.Errorf("Unknown attribute %s", key)
	return
}

func (b *Bilbo) RenderPage(page *Page) (err error) {
	// Render the page according to its extension
	ext := path.Ext(page.Filepath)
	switch ext {
	case ".md", ".markdown":
		extensions := blackfriday.NoIntraEmphasis | blackfriday.Tables | blackfriday.FencedCode | blackfriday.Strikethrough | blackfriday.SpaceHeadings | blackfriday.AutoHeadingIDs | blackfriday.HeadingIDs | blackfriday.BackslashLineBreak | blackfriday.DefinitionLists
		page.Rendered = blackfriday.Run(page.Source, blackfriday.WithExtensions(extensions))
	default:
		page.Rendered = page.Source
	}

	doc, err := html.Parse(bytes.NewReader(page.Rendered))
	if err != nil {
		return
	}

	firstHeadlineFound := false

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && !firstHeadlineFound && n.DataAtom == atom.H1 {
			// Extract text from first headline node
			firstHeadlineFound = true
			page.Title = htmlNodeExtractString(n)
		}

		if n.Type == html.ElementNode && headlineRegex.MatchString(n.Data) {
			// Prepend anchor link to headlines
			id, err := htmlNodeGetAttributeByKey(n, "id")
			if err != nil {
				return
			}

			n.Attr = append(n.Attr, html.Attribute{Key: "class", Val: "a-hl"})

			linkNode := &html.Node{
				Type:     html.ElementNode,
				DataAtom: atom.A,
				Data:     "a",
				Attr: []html.Attribute{
					{Key: "class", Val: "a-hl__anchor"},
					{Key: "href", Val: "#" + id},
					{Key: "title", Val: "Set anchor to \"" + htmlNodeExtractString(n) + "\""},
				},
			}
			linkNode.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: "🔗",
			})

			n.InsertBefore(linkNode, n.FirstChild)
		}

		if n.Type == html.TextNode {
			// Gollum-Style link tags https://github.com/gollum/gollum/wiki#link-tag
			match := linkRegex.FindAllStringSubmatch(n.Data, 3)
			if match != nil {
				hasTitle := match[0][3] == ""

				rawTitle := match[0][1]
				if !hasTitle {
					rawTitle = match[0][3]
				}

				rawLink := match[0][2]
				if !hasTitle {
					rawLink = match[0][3]
				}

				link, err := url.Parse(rawLink)
				if err != nil {
					return
				}

				classes := "external"
				if !link.IsAbs() {
					classes = "internal"
					link.Path = normalizePageLink(link.Path, true)

					if _, err := b.getPage(link.Path, false); err != nil {
						classes = "internal absent"
					}
				}

				linkNode := &html.Node{
					Type:     html.ElementNode,
					DataAtom: atom.A,
					Data:     "a",
					Attr: []html.Attribute{
						{Key: "class", Val: classes},
						{Key: "href", Val: link.String()},
					},
				}
				linkNode.AppendChild(&html.Node{
					Type: html.TextNode,
					Data: rawTitle,
				})

				n.Parent.InsertBefore(linkNode, n)
				n.Parent.RemoveChild(n)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, doc)
	page.Rendered = buf.Bytes()

	return
}
