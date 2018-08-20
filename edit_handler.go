package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"net/http"
	"path"
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
		if err == io.EOF {
			page = &Page{
				Filepath: vars["page"],
				Title:    normalizePageLink(vars["page"], false),
			}
		} else {
			panic(err)
		}
	}

	commitMsg := fmt.Sprintf("Updated %s", page.Title)
	if len(page.Source) == 0 {
		commitMsg = fmt.Sprintf("Created %s", page.Title)
	}

	b.renderTemplate(w, r, "edit.html", hash{
		"commitMsg":  commitMsg,
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

func (b *Bilbo) HandleEditNew(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		folder := r.PostFormValue("page-folder")
		name := r.PostFormValue("page-name")
		if !path.IsAbs(name) {
			name = path.Join(folder, name)
		}
		name = normalizePageLink(name, true)

		http.Redirect(w, r, path.Join("/edit/", name), http.StatusMovedPermanently)
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
