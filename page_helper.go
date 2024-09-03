package main

import (
	"net/http"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gorilla/mux"
)

// Retrieve the requested page
func (b *Bilbo) getRequestedPage(r *http.Request) (page *Page, commit plumbing.Hash, err error) {
	commit = r.Context().Value("GitHead").(plumbing.Hash)

	vars := mux.Vars(r)
	pageLink := normalizePageLink(vars["page"], true)

	page, err = b.getPageAtCommit(pageLink, true, commit)

	return
}
