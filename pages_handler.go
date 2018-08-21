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

	linkpath, err := mux.Get("pagesIndex").URL()
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
		linkpath, err = mux.Get("pages").URL("folder", path.Join(currentPrefix, folder))
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

	commit := r.Context().Value("GitHead").(plumbing.Hash)
	folderStructure, err := b.getPagesAtCommit(vars["folder"], commit)
	if err != nil {
		panic(err)
	}

	pageTitle := path.Clean(vars["folder"])
	if pageTitle == "." {
		pageTitle = "all pages"
	}

	breadcrumb, err := createBreadcrumb(b.mux, vars["folder"])
	if err != nil {
		panic(err)
	}

	pageFolder := path.Clean(path.Join("/", vars["folder"]))
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
