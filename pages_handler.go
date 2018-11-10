package main

import (
	"github.com/gorilla/mux"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"net/http"
	"path"
	"strings"
)

func createBreadcrumb(mux *mux.Router, folderPath string) (hierarchy []*Folder, err error) {
	currentPrefix := "/"

	linkpath, err := mux.Get("pages#index").URL("folder", "")
	if err != nil {
		panic(err)
	}

	hierarchy = append(hierarchy, &Folder{
		Title:    "Home",
		Linkpath: linkpath.String(),
	})

	if folderPath == "" {
		return
	}

	for _, folder := range strings.Split(folderPath, "/") {
		linkpath, err = mux.Get("pages#index").URL("folder", path.Join(currentPrefix, folder))
		if err != nil {
			panic(err)
		}

		hierarchy = append(hierarchy, &Folder{
			Title:    folder,
			Linkpath: linkpath.String(),
		})

		currentPrefix = path.Join(currentPrefix, folder)
	}

	return
}

func (b *Bilbo) HandlePages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	folder := ""
	if len(vars["folder"]) > 0 {
		folder = vars["folder"]
	}

	commit := r.Context().Value("GitHead").(plumbing.Hash)
	folderStructure, err := b.getPagesAtCommit(folder, commit)
	if err != nil {
		panic(err)
	}

	pageTitle := path.Clean(folder)
	if pageTitle == "." {
		pageTitle = "all pages"
	}

	breadcrumb, err := createBreadcrumb(b.mux, folder)
	if err != nil {
		panic(err)
	}

	pageFolder := path.Clean(path.Join("/", folder))
	if pageFolder == "/" {
		pageFolder = ""
	}

	b.renderTemplate(w, r, "pages.html", hash{
		"breadcrumb":  breadcrumb,
		"directories": folderStructure.Folders,
		"pageFolder":  pageFolder,
		"pageLayout":  "pages",
		"pages":       folderStructure.Pages,
		"pageTitle":   pageTitle,
	})
}
