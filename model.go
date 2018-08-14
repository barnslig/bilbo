package main

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io"
	"io/ioutil"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func (b *Bilbo) getPageHistory(fileName string, onlyLast bool) (commits []*object.Commit, err error) {
	head, err := b.repo.Head()
	if err != nil {
		return
	}

	return b.getPageHistoryFromCommit(fileName, onlyLast, head.Hash())
}

func (b *Bilbo) getPageHistoryFromCommit(fileName string, onlyLast bool, commit plumbing.Hash) (commits []*object.Commit, err error) {
	ci, err := b.repo.Log(&git.LogOptions{})
	if err != nil {
		return
	}

	var (
		currentCommit     *object.Commit
		currentCommitFile *object.File
		lastCommit        *object.Commit
		lastCommitFile    *object.File
	)

	lastCommit, err = ci.Next()
	if err != nil {
		return
	}

	for {
		currentCommit, err = ci.Next()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}

			return
		}

		lastCommitFile, err = lastCommit.File(fileName)
		if err != nil && err != object.ErrFileNotFound {
			break
		}

		currentCommitFile, err = currentCommit.File(fileName)
		if err != nil && err != object.ErrFileNotFound {
			break
		}

		err = nil

		if (lastCommitFile == nil && currentCommitFile != nil) ||
			(lastCommitFile != nil && currentCommitFile == nil) ||
			(lastCommitFile != nil && currentCommitFile != nil && lastCommitFile.Blob.Hash != currentCommitFile.Blob.Hash) {
			commits = append(commits, lastCommit)

			if onlyLast {
				return
			}
		}

		lastCommit = currentCommit
	}

	return
}

func (b *Bilbo) getPage(fileName string, withLastCommit bool) (page *Page, err error) {
	head, err := b.repo.Head()
	if err != nil {
		return
	}

	return b.getPageAtCommit(fileName, withLastCommit, head.Hash())
}

func (b *Bilbo) getPageAtCommit(fileName string, withLastCommit bool, commit plumbing.Hash) (page *Page, err error) {
	obj, err := b.repo.CommitObject(commit)
	if err != nil {
		return
	}

	iter, err := obj.Files()
	if err != nil {
		return
	}
	defer iter.Close()

	cleanFileName := normalizePageLink(strings.TrimSuffix(fileName, path.Ext(fileName)), false)

	var (
		file     *object.File
		fileLink string
	)
	for {
		file, err = iter.Next()
		if err != nil {
			return
		}

		// match without file extension
		fileLink = strings.TrimSuffix(file.Name, path.Ext(file.Name))
		if fileLink == cleanFileName {
			break
		}
	}

	// Read page into byte slice
	reader, err := file.Blob.Reader()
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return
	}

	// Get file history
	var (
		history    []*object.Commit
		lastCommit *object.Commit
	)
	if withLastCommit {
		history, err = b.getPageHistory(file.Name, true)
		if err != nil {
			return
		}
		lastCommit = history[0]
	}

	// Create page
	page = &Page{
		LastCommit: lastCommit,
		Filepath:   file.Name,
		Linkpath:   normalizePageLink(fileLink, true),
		Source:     data,
		Title:      path.Base(fileLink),
	}

	return
}

func (b *Bilbo) getPages(folderPath string) (directories []*Folder, pages []*Page, err error) {
	head, err := b.repo.Head()
	if err != nil {
		return
	}

	return b.getPagesAtCommit(folderPath, head.Hash())
}

func (b *Bilbo) getPagesAtCommit(folderPath string, commit plumbing.Hash) (directories []*Folder, pages []*Page, err error) {
	obj, err := b.repo.CommitObject(commit)
	if err != nil {
		return
	}

	tree, err := obj.Tree()
	if err != nil {
		return
	}

	cleanFolderPath := path.Clean(folderPath)

	if cleanFolderPath != "." {
		directories = append(directories, &Folder{
			Title:    "..",
			Linkpath: "..",
		})

		tree, err = tree.Tree(cleanFolderPath)
		if err != nil {
			return
		}
	}

	var (
		linkpath *url.URL
		page     *Page
	)
	for _, entry := range tree.Entries {
		if entry.Mode == filemode.Dir {
			linkpath, err = b.mux.Get("pages").URL("folder", path.Join(cleanFolderPath, entry.Name))
			if err != nil {
				return
			}

			directories = append(directories, &Folder{
				Title:    entry.Name,
				Linkpath: linkpath.String(),
			})
		} else {
			page, err = b.getPageAtCommit(path.Join(cleanFolderPath, entry.Name), false, commit)
			if err != nil {
				return
			}
			pages = append(pages, page)
		}
	}

	return
}

func (b *Bilbo) updatePage(fileName string, data string, message string) (err error) {
	pageFilePath := fileName

	// Make sure the file has an extension
	// TODO file type switch
	ext := filepath.Ext(pageFilePath)
	if ext != ".md" && ext != ".markdown" {
		pageFilePath = fmt.Sprintf("%s.%s", pageFilePath, "md")
	}

	// Make a safe file path
	base, err := filepath.Abs(b.cfg.DataDir)
	if err != nil {
		return
	}

	orig, err := filepath.Abs(filepath.Join(b.cfg.DataDir, pageFilePath))
	if err != nil {
		return
	}

	if !strings.HasPrefix(orig, base) {
		err = fmt.Errorf("Path breaks out of data directory")
		return
	}

	// Write to file
	// TODO configurable permission mode
	err = ioutil.WriteFile(orig, []byte(data), 0644)
	if err != nil {
		return
	}

	// Stage file
	wt, err := b.repo.Worktree()
	if err != nil {
		return
	}

	_, err = wt.Add(pageFilePath)
	if err != nil {
		return
	}

	// Create commit
	// TODO author from config
	_, err = wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Anonymous",
			Email: "anon@anon.com",
			When:  time.Now(),
		},
	})

	return
}
