package main

import (
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type Page struct {
	LastCommit *object.Commit
	Filepath   string
	Linkpath   string
	Rendered   []byte
	Source     []byte
	Title      string
}
