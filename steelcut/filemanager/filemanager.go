package filemanager

import (
	"os"
	"time"
)

// DirOperations represents operations that can be performed on directories.
type DirOperations interface {
	CreateDirectory(path string) error
	DeleteDirectory(path string) error
	MoveDirectory(sourcePath, destPath string) error
	CopyDirectory(sourcePath, destPath string) error
	ListDirectory(path string) ([]string, error)
	GetDirAttributes(path string) (Directory, error)
	DiskUsage(path string) (DiskUsageInfo, error)
}

// FileOperations represents operations that can be performed on files.
type FileOperations interface {
	CreateFile(path string) error
	DeleteFile(path string) error
	MoveFile(sourcePath, destPath string) error
	CopyFile(sourcePath, destPath string) error
	GetFileAttributes(path string) (File, error)
}

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

type DiskUsageInfo struct {
	Total      int64   // total space in bytes
	Used       int64   // used space in bytes
	Available  int64   // available space in bytes
	UsePercent float64 // usage percentage
}

// Directory describes basic directory attributes.
type Directory struct {
	Path     string
	Mode     os.FileMode
	Modified time.Time
}
