package main

import (
	"encoding/json"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"net/http"
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

	page := &Page{
		Filepath: data.Filepath,
		Source:   []byte(data.Data),
	}

	commit := r.Context().Value("GitHead").(plumbing.Hash)
	err = b.RenderPage(page, commit)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(page.Rendered)
}
