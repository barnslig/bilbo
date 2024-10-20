package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/patrickmn/go-cache"
)

func (b *Bilbo) getPageHistoryFromCommit(fileName string, onlyLast bool, commit plumbing.Hash) (commits []*object.Commit, err error) {
	// Determine cache key
	cacheKey := fmt.Sprintf("getPageHistoryFromCommit-%s-%t-%s", fileName, onlyLast, commit)

	// Try to retrieve page from cache
	cachedCommits, found := b.cache.Get(cacheKey)
	if found {
		commits = cachedCommits.([]*object.Commit)
		return
	}

	ci, err := b.repo.Log(&git.LogOptions{
		From:     commit,
		FileName: &fileName,
	})
	if err != nil {
		return
	}

	var currentCommit *object.Commit

	for {
		currentCommit, err = ci.Next()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}

			return
		}

		commits = append(commits, currentCommit)

		if onlyLast {
			break
		}
	}

	b.cache.Set(cacheKey, commits, cache.DefaultExpiration)
	return
}

func (b *Bilbo) getPageAtCommit(fileName string, withLastCommit bool, commit plumbing.Hash) (page *Page, err error) {
	// Determine cache key
	cacheKey := fmt.Sprintf("getPageAtCommit-%s-%t-%s", fileName, withLastCommit, commit)

	// Try to retrieve page from cache
	cachedPage, found := b.cache.Get(cacheKey)
	if found {
		page = cachedPage.(*Page)
		return
	}

	obj, err := b.repo.CommitObject(commit)
	if err != nil {
		return
	}

	iter, err := obj.Files()
	if err != nil {
		return
	}
	defer iter.Close()

	cleanFileName := normalizePageLink(fileName, false)

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

	data, err := io.ReadAll(reader)
	if err != nil {
		return
	}

	// Get file history
	var (
		history    []*object.Commit
		lastCommit *object.Commit
	)
	if withLastCommit {
		history, err = b.getPageHistoryFromCommit(file.Name, true, commit)
		if err != nil {
			return
		}
		lastCommit = history[0]
	}

	// Create page
	page = &Page{
		Commit:     obj,
		LastCommit: lastCommit,
		Filepath:   file.Name,
		Linkpath:   normalizePageLink(fileLink, true),
		Source:     data,
		Title:      path.Base(fileLink),
	}

	// Cache page
	b.cache.Set(cacheKey, page, cache.DefaultExpiration)

	return
}

func (b *Bilbo) getPagesAtCommit(folderPath string, commit plumbing.Hash) (folderStructure *FolderStructure, err error) {
	folderStructure = &FolderStructure{}

	// Determine cache key
	cacheKey := fmt.Sprintf("getPagesAtCommit-%s-%s", folderPath, commit)

	// Try to retrieve directories and pages from cache
	cachedFolderStructure, found := b.cache.Get(cacheKey)
	if found {
		folderStructure = cachedFolderStructure.(*FolderStructure)
		return
	}

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
		folderStructure.Folders = append(folderStructure.Folders, &Folder{
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
			linkpath, err = b.mux.Get("pages#index").URL("folder", path.Join(cleanFolderPath, entry.Name))
			if err != nil {
				return
			}

			folderStructure.Folders = append(folderStructure.Folders, &Folder{
				Title:    entry.Name,
				Linkpath: linkpath.String(),
			})
		} else {
			page, err = b.getPageAtCommit(path.Join(cleanFolderPath, entry.Name), false, commit)
			if err != nil {
				return
			}
			folderStructure.Pages = append(folderStructure.Pages, page)
		}
	}

	// Cache folder structure
	b.cache.Set(cacheKey, folderStructure, cache.DefaultExpiration)

	return
}

func (b *Bilbo) updatePage(fileName string, data string, message string, author GitAuthor) (err error) {
	// Make sure the file has an extension
	pageFilePath := ensureFileNameHasExtension(fileName)

	// Make sure the file stays within our data dir
	absPageFilePath, err := ensureSaveFilePath(pageFilePath, b.cfg.DataDir, true)
	if err != nil {
		return
	}

	// Write to file
	// TODO configurable permission mode
	err = os.WriteFile(absPageFilePath, []byte(data), 0644)
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
	_, err = wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  author.Name,
			Email: author.Email,
			When:  time.Now(),
		},
	})

	return
}

func (b *Bilbo) movePage(currentFileName string, nextFileName string, message string, author GitAuthor) (err error) {
	// Make sure both files are within our data dir
	absCurrentFileName, err := ensureSaveFilePath(currentFileName, b.cfg.DataDir, false)
	if err != nil {
		return
	}

	absNextFileName, err := ensureSaveFilePath(nextFileName, b.cfg.DataDir, false)
	if err != nil {
		return
	}

	// Make sure the new file has an extension
	absNextFileName = ensureFileNameHasExtension(absNextFileName)

	// Just try to move the file
	wt, err := b.repo.Worktree()
	if err != nil {
		return
	}

	_, err = wt.Move(strings.TrimPrefix(absCurrentFileName, "/"), strings.TrimPrefix(absNextFileName, "/"))
	if err != nil {
		return
	}

	// Create commit
	_, err = wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  author.Name,
			Email: author.Email,
			When:  time.Now(),
		},
	})

	return
}

func (b *Bilbo) deletePage(fileName string, message string, author GitAuthor) (err error) {
	// Make sure file is within our data dir
	absFileName, err := ensureSaveFilePath(fileName, b.cfg.DataDir, false)
	if err != nil {
		return
	}

	// Just try to delete the file
	wt, err := b.repo.Worktree()
	if err != nil {
		return
	}

	_, err = wt.Remove(strings.TrimPrefix(absFileName, "/"))
	if err != nil {
		return
	}

	// Create commit
	_, err = wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  author.Name,
			Email: author.Email,
			When:  time.Now(),
		},
	})

	return
}
