package filemanager

// FileOperations represents operations that can be performed on files.
type FileOperations interface {
	CreateFile(path string) error
	DeleteFile(path string) error
	MoveFile(sourcePath, destPath string) error
	CopyFile(sourcePath, destPath string) error
	GetFileAttributes(path string) (File, error)
}
