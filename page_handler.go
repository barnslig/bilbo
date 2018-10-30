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
		redirectUrl, err := b.mux.Get("page").URL("page", normalizedLink)
		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, redirectUrl.String(), http.StatusMovedPermanently)
	}

	page, commit, err := b.getRequestedPage(r)
	if err != nil {
		panic(err)
	}

	err = b.RenderPage(page, commit)
	if err != nil {
		panic(err)
	}

	pageFolder := path.Dir(page.Linkpath)
	if pageFolder == "/" {
		pageFolder = ""
	}

	b.renderTemplate(w, r, "page.html", hash{
		"content":    template.HTML(string(page.Rendered)),
		"isPage":     true,
		"lastCommit": page.LastCommit,
		"page":       page,
		"pageFolder": pageFolder,
		"pageLayout": "page",
		"pageTitle":  page.Title,
	})
}
