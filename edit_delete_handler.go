package main

import (
	"fmt"
	"net/http"
	"path"
)

func (b *Bilbo) HandleEditDelete(w http.ResponseWriter, r *http.Request) {
	page, _, err := b.getRequestedPage(r)
	if err != nil {
		panic(err)
	}

	if r.Method == "POST" {
		err = b.deletePage(page.Filepath, fmt.Sprintf("Deleted %s", path.Base(page.Filepath)))
		if err != nil {
			panic(err)
		}

		redirectUrl, err := b.mux.Get("pages#new").URL("folder", "")
		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, redirectUrl.String(), http.StatusSeeOther)
		return
	}

	b.renderTemplate(w, r, "edit_delete.html", hash{
		"page":       page,
		"pageLayout": "form",
		"pageTitle":  fmt.Sprintf("Delete page \"%s\"", page.Title),
	})
}
