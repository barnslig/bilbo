package main

import (
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Page struct {
	LastCommit *object.Commit
	Filepath   string
	Linkpath   string
	Rendered   []byte
	Source     []byte
	Title      string
}
