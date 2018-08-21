package main

import (
	"github.com/gorilla/mux"
	"gopkg.in/src-d/go-git.v4/plumbing"
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

	commit := r.Context().Value("GitHead").(plumbing.Hash)

	page, err := b.getPageAtCommit(vars["page"], true, commit)
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
