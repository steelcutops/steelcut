package filemanager

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

// FileOperations represents operations that can be performed on files.
type FileOperations interface {
	CreateFile(path string) error
	DeleteFile(path string) error
	MoveFile(sourcePath, destPath string) error
	CopyFile(sourcePath, destPath string) error
	GetFileAttributes(path string) (File, error)
}

// FileOperations methods

func (f *FileManagerImpl) CreateFile(path string) error {
	config := cm.CommandConfig{
		Command: "touch",
		Args:    []string{path},
	}
	result, err := f.commandManager.Run(context.TODO(), f.host, config)
	return handleCommandResult(result, err)
}

func (f *FileManagerImpl) DeleteFile(path string) error {
	config := cm.CommandConfig{
		Command: "rm",
		Args:    []string{path},
	}
	result, err := f.commandManager.Run(context.TODO(), f.host, config)
	return handleCommandResult(result, err)
}

func (f *FileManagerImpl) MoveFile(sourcePath, destPath string) error {
	config := cm.CommandConfig{
		Command: "mv",
		Args:    []string{sourcePath, destPath},
	}
	result, err := f.commandManager.Run(context.TODO(), f.host, config)
	return handleCommandResult(result, err)
}

func (f *FileManagerImpl) CopyFile(sourcePath, destPath string) error {
	config := cm.CommandConfig{
		Command: "cp",
		Args:    []string{sourcePath, destPath},
	}
	result, err := f.commandManager.Run(context.TODO(), f.host, config)
	return handleCommandResult(result, err)
}

func (f *FileManagerImpl) GetFileAttributes(path string) (File, error) {
	config := cm.CommandConfig{
		Command: "stat",
		Args:    []string{"-c", "%s %F %Y", path}, // Get size, file type, and modification time
	}
	result, err := f.commandManager.Run(context.TODO(), f.host, config)
	if err != nil {
		return File{}, err
	}

	// Split the output
	parts := strings.Split(strings.TrimSpace(result.STDOUT), " ")

	if len(parts) != 3 {
		return File{}, fmt.Errorf("unexpected stat output format: %s", result.STDOUT)
	}

	// Extract size
	size, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return File{}, fmt.Errorf("error parsing file size: %v", err)
	}

	// Extract FileMode - this is a simplification and may not represent the exact mode
	var mode os.FileMode
	if parts[1] == "regular file" {
		mode = 0644 // This is a simplification. Actual permissions might vary.
	} else if parts[1] == "directory" {
		mode = 0755 // Again, a simplification.
	} else {
		mode = 0000
	}

	// Extract modification time
	modifiedSeconds, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return File{}, fmt.Errorf("error parsing modification time: %v", err)
	}
	modified := time.Unix(modifiedSeconds, 0)

	return File{
		Path:     path,
		Size:     size,
		Mode:     mode,
		Modified: modified,
	}, nil
}
