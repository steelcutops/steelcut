package filemanager

// DirOperations represents operations that can be performed on directories.
type DirOperations interface {
	CreateDirectory(path string) error
	DeleteDirectory(path string) error
	MoveDirectory(sourcePath, destPath string) error
	CopyDirectory(sourcePath, destPath string) error
	ListDirectory(path string) ([]string, error)
	GetDirAttributes(path string) (Directory, error)
}
