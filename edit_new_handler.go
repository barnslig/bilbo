package main

import (
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

func (b *Bilbo) HandleEditNew(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if r.Method == "POST" {
		name := r.FormValue("page-name")
		if !path.IsAbs(name) {
			name = path.Join(vars["folder"], name)
		}
		name = normalizePageLink(name, true)

		redirectUrl, err := b.mux.Get("pages#edit").URL("page", name)
		if err != nil {
			panic(err)
		}

		http.Redirect(w, r, redirectUrl.String(), http.StatusSeeOther)
		return
	}

	folder := normalizePageLink(vars["folder"], true)

	b.renderTemplate(w, r, "edit_new.html", hash{
		"folder":     folder,
		"pageLayout": "form",
		"pageTitle":  "Create new page",
	})
}
