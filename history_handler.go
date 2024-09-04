package main

import (
	"fmt"
	"net/http"
	"path"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gorilla/mux"
)

func (b *Bilbo) HandleHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	normalizedLink := normalizePageLink(vars["page"], false)
	if normalizedLink != vars["page"] {
		redirectUrl, err := b.mux.Get("pages#history").URL("page", normalizedLink)
		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, redirectUrl.String(), http.StatusMovedPermanently)
	}

	commit := r.Context().Value(key("GitHead")).(plumbing.Hash)
	pageHistory, err := b.getPageHistoryFromCommit(vars["page"]+".md", false, commit)
	if err != nil {
		panic(err)
	}

	page, _, err := b.getRequestedPage(r)
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
