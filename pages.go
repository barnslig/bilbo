package main

type Folder struct {
	Title    string
	Linkpath string
}

type FolderStructure struct {
	Folders []*Folder
	Pages   []*Page
}
