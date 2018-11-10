package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"path"
)

func (b *Bilbo) HandleEditNew(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		folder := r.FormValue("page-folder")

		name := r.FormValue("page-name")
		if !path.IsAbs(name) {
			name = path.Join(folder, name)
		}
		name = normalizePageLink(name, true)

		redirectUrl, err := b.mux.Get("pages#edit").URL("page", name)
		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, redirectUrl.String(), http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	folder := normalizePageLink(vars["folder"], true)

	b.renderTemplate(w, r, "edit_new.html", hash{
		"folder":     folder,
		"pageLayout": "form",
		"pageTitle":  "Create new page",
	})
}
