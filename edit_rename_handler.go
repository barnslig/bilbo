package main

import (
	"fmt"
	"net/http"
	"path"
)

func (b *Bilbo) HandleEditRename(w http.ResponseWriter, r *http.Request) {
	currentPage, _, err := b.getRequestedPage(r)
	if err != nil {
		panic(err)
	}

	if r.Method == "POST" {
		// Make sure next page linkpath is absolute
		nextPageLinkpath := r.FormValue("next-page-name")
		if !path.IsAbs(nextPageLinkpath) {
			nextPageLinkpath = path.Join(path.Dir(currentPage.Linkpath), nextPageLinkpath)
		}
		nextPageLinkpath = normalizePageLink(nextPageLinkpath, false)

		// Make sure we keep the file extension
		nextPageFilepath := nextPageLinkpath
		if path.Ext(nextPageFilepath) == "" && path.Ext(currentPage.Filepath) != "" {
			nextPageFilepath = nextPageFilepath + path.Ext(currentPage.Filepath)
		}

		err = b.movePage(currentPage.Filepath, nextPageFilepath, fmt.Sprintf("Renamed %s", path.Base(nextPageFilepath)))
		if err != nil {
			panic(err)
		}

		redirectUrl, err := b.mux.Get("pages#show").URL("page", nextPageLinkpath)
		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, redirectUrl.String(), http.StatusSeeOther)
		return
	}

	b.renderTemplate(w, r, "edit_rename.html", hash{
		"currentFolder": path.Dir(currentPage.Linkpath),
		"currentPage":   currentPage,
		"pageLayout":    "form",
		"pageTitle":     fmt.Sprintf("Rename page \"%s\"", currentPage.Title),
	})
}
