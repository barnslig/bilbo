package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Make sure a filename has an extension
func ensureFileNameHasExtension(_fileName string) (fileName string) {
	fileName = _fileName

	// TODO file type switch
	ext := filepath.Ext(fileName)
	if ext != ".md" && ext != ".markdown" {
		fileName = fmt.Sprintf("%s.%s", fileName, "md")
	}

	return
}

// Check whether filePath stays within basePath
func ensureSaveFilePath(filePath string, basePath string, returnAbs bool) (savePath string, err error) {
	// Check that basePath can be turned into an absolute path
	base, err := filepath.Abs(basePath)
	if err != nil {
		return
	}

	// Do nothing when we already have a save filePath
	absFilePath, err := filepath.Abs(filePath)
	if err == nil && strings.HasPrefix(absFilePath, base) {
		return
	}

	// Check that basePath joined with filePath can be turned into an absolute path
	absPath, err := filepath.Abs(filepath.Join(basePath, filePath))
	if err != nil {
		return
	}

	// Check that the joined path still has the basePath as prefix
	if !strings.HasPrefix(absPath, base) {
		err = fmt.Errorf("Path breaks out of data directory")
		return
	}

	// Should we return an absolute path?
	if returnAbs {
		savePath = absPath
		return
	}

	savePath = strings.TrimPrefix(absPath, base)
	return
}
