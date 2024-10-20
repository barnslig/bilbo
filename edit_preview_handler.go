package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing"
)

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

	commit := r.Context().Value(key("GitHead")).(plumbing.Hash)
	commitObj, err := b.repo.CommitObject(commit)
	if err != nil {
		return
	}

	page := &Page{
		Commit:   commitObj,
		Filepath: data.Filepath,
		Source:   []byte(data.Data),
	}

	err = b.RenderPage(page, false)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(page.Rendered)
}
