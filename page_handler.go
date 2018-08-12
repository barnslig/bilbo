package main

import (
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"strings"
)

func (b *Bilbo) HandlePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Make sure we have no spaces; replace them by dashes
	if strings.Contains(vars["page"], " ") {
		http.Redirect(w, r, normalizePageLink(vars["page"], true), http.StatusMovedPermanently)
	}

	page, err := b.getPage(vars["page"], true)
	if err != nil {
		panic(err)
	}

	err = b.RenderPage(page)
	if err != nil {
		panic(err)
	}

	b.renderTemplate(w, r, "page.html", map[string]interface{}{
		"lastCommit": page.LastCommit,
		"content":    template.HTML(string(page.Rendered)),
		"pageTitle":  page.Title,
	})
}
