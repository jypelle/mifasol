package tool

import (
	"io/fs"
	"os"
	"time"
)

type StaticFileInfoWrapper struct {
	os.FileInfo
	fixedModTime time.Time
}

func (f *StaticFileInfoWrapper) ModTime() time.Time {
	return f.fixedModTime
}

type StaticFSWrapper struct {
	fs.ReadDirFS
	FixedModTime time.Time
}

func (f *StaticFSWrapper) Open(name string) (fs.File, error) {
	file, err := f.ReadDirFS.Open(name)

	return &StaticFileWrapper{File: file, fixedModTime: f.FixedModTime}, err
}

func (f *StaticFSWrapper) ReadDir(name string) ([]fs.DirEntry, error) {
	return f.ReadDirFS.ReadDir(name)
}

type StaticFileWrapper struct {
	fs.File
	fixedModTime time.Time
}

func (f *StaticFileWrapper) Stat() (os.FileInfo, error) {

	fileInfo, err := f.File.Stat()

	return &StaticFileInfoWrapper{FileInfo: fileInfo, fixedModTime: f.fixedModTime}, err
}
