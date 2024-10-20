package main

import (
	"path"
)

type PageRenderer interface {
	Render(page *Page) error
}

func (b *Bilbo) RenderPage(page *Page) (err error) {
	// Render the page according to its extension
	ext := path.Ext(page.Filepath)
	switch ext {
	case ".md", ".markdown":
		r := NewMarkdownPageRenderer(MarkdownPageRendererOptions{
			pageExists: func(path string) bool {
				_, err := b.getPageAtCommit(path, false, page.Commit.Hash)
				return err == nil
			},
		})
		err = r.Render(page)
	default:
		page.Rendered = page.Source
	}

	return
}
