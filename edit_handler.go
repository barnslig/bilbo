package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
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

	page, err := b.getPage(vars["page"], true)
	if err != nil {
		panic(err)
	}

	b.renderTemplate(w, r, "edit.html", hash{
		"page":       page,
		"pageLayout": "editor",
		"pageTitle":  fmt.Sprintf("Edit page \"%s\"", page.Title),
		"source":     template.HTML(string(page.Source)),
	})
}

type EditPreviewData struct {
	Data     string `json:"data"`
	Filepath string `json:"filepath"`
}

func (b *Bilbo) HandleEditPreview(w http.ResponseWriter, r *http.Request) {
	data := EditPreviewData{}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		panic(err)
	}

	page := &Page{
		Filepath: data.Filepath,
		Source:   []byte(data.Data),
	}

	err = b.RenderPage(page)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(page.Rendered)
}
