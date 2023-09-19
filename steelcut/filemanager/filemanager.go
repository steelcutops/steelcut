package filemanager

import (
	"errors"
	"os"
	"time"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
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

type FileManagerImpl struct {
	commandManager cm.CommandManager
}

func NewFileManager(commandManager cm.CommandManager) *FileManagerImpl {
	return &FileManagerImpl{
		commandManager: commandManager,
	}
}

func handleCommandResult(result cm.CommandResult, err error) error {
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return errors.New(result.STDERR)
	}
	return nil
}
