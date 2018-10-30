package main

import (
	"github.com/gorilla/mux"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"net/http"
)

// Retrieve the requested page
func (b *Bilbo) getRequestedPage(r *http.Request) (page *Page, commit plumbing.Hash, err error) {
	commit = r.Context().Value("GitHead").(plumbing.Hash)

	vars := mux.Vars(r)
	pageLink := normalizePageLink(vars["page"], true)

	page, err = b.getPageAtCommit(pageLink, true, commit)

	return
}
