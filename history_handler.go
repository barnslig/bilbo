package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"net/http"
	"path"
)

func (b *Bilbo) HandleHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	normalizedLink := normalizePageLink(vars["page"], false)
	if normalizedLink != vars["page"] {
		redirectUrl, err := b.mux.Get("history").URL("page", normalizedLink)
		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, redirectUrl.String(), http.StatusMovedPermanently)
	}

	commit := r.Context().Value("GitHead").(plumbing.Hash)
	pageHistory, err := b.getPageHistoryFromCommit(vars["page"]+".md", false, commit)
	if err != nil {
		panic(err)
	}

	page, err := b.getPageAtCommit(vars["page"], false, commit)
	if err != nil {
		panic(err)
	}

	pageFolder := path.Dir(page.Linkpath)
	if pageFolder == "/" {
		pageFolder = ""
	}

	b.renderTemplate(w, r, "history.html", hash{
		"isPage":      true,
		"page":        page,
		"pageFolder":  pageFolder,
		"pageHistory": pageHistory,
		"pageLayout":  "commits",
		"pageTitle":   fmt.Sprintf("History for %s", page.Title),
	})
}
