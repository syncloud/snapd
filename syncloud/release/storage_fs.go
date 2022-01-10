package main

import (
	"github.com/otiai10/copy"
	"io/ioutil"
	"path"
)

type FileSystem struct {
	target string
}

func NewFileSystem(target string) *FileSystem {
	return &FileSystem{target: target}
}

func (f *FileSystem) UploadFile(from string, to string) error {
	return copy.Copy(from, path.Join(f.target, to))
}

func (f *FileSystem) UploadContent(content string, to string) error {
	return ioutil.WriteFile(path.Join(f.target, to), []byte(content), 0644)
}
