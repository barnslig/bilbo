package main

import (
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"path"
)

func (b *Bilbo) HandlePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	normalizedLink := normalizePageLink(vars["page"], false)
	if normalizedLink != vars["page"] {
		http.Redirect(w, r, path.Join("/", normalizedLink), http.StatusMovedPermanently)
	}

	page, err := b.getPage(vars["page"], true)
	if err != nil {
		panic(err)
	}

	err = b.RenderPage(page)
	if err != nil {
		panic(err)
	}

	b.renderTemplate(w, r, "page.html", hash{
		"content":    template.HTML(string(page.Rendered)),
		"isPage":     true,
		"lastCommit": page.LastCommit,
		"page":       page,
		"pageLayout": "page",
		"pageTitle":  page.Title,
	})
}
