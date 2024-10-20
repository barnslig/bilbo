package main

import (
	"fmt"
	"path"

	"github.com/patrickmn/go-cache"
)

type PageRenderer interface {
	Render(page *Page) error
}

func (b *Bilbo) RenderPage(page *Page, enableCache bool) (err error) {
	cacheKey := fmt.Sprintf("page-%s-%s", page.Filepath, page.Commit.Hash)

	if enableCache {
		c, found := b.cache.Get(cacheKey)
		if found {
			cachedPage := c.(*Page)
			page.Title = cachedPage.Title
			page.Rendered = cachedPage.Rendered
			return
		}
	}

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

	if enableCache {
		b.cache.Set(cacheKey, page, cache.DefaultExpiration)
	}

	return
}
