package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"net/http"
)

type EditData struct {
	Data    string `json:"data"`
	Message string `json:"message"`
}

func (b *Bilbo) HandleEdit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if r.Method == "POST" {
		data := EditData{}

		// Parse request data
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			panic(err)
		}

		err = b.updatePage(vars["page"], data.Data, data.Message)
		if err != nil {
			panic(err)
		}

		w.Write([]byte("ok"))
		return
	}

	isNewPage := false
	page, _, err := b.getRequestedPage(r)
	if err != nil {
		if err == io.EOF {
			isNewPage = true
			page = &Page{
				Filepath: fmt.Sprintf("%s.md", vars["page"]),
				Linkpath: vars["page"],
				Title:    normalizePageLink(vars["page"], false),
			}
		} else {
			panic(err)
		}
	}

	commitMsg := fmt.Sprintf("Update %s", page.Title)
	if len(page.Source) == 0 {
		commitMsg = fmt.Sprintf("Create %s", page.Title)
	}

	b.renderTemplate(w, r, "edit.html", hash{
		"commitMsg":  commitMsg,
		"isNewPage":  isNewPage,
		"page":       page,
		"pageLayout": "editor",
		"pageTitle":  commitMsg,
		"source":     template.HTML(string(page.Source)),
	})
}
