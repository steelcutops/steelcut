package filemanager

import (
	"os"
	"time"
)

// FileManager encompasses operations on both files and directories.
type FileManager interface {
	FileOperations
	DirOperations
}

// File describes basic file attributes.
type File struct {
	Path     string
	Size     int64 // bytes
	Mode     os.FileMode
	Modified time.Time
}

// Directory describes basic directory attributes.
type Directory struct {
	Path     string
	Mode     os.FileMode
	Modified time.Time
}
