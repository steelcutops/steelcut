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

// DirOperations represents operations that can be performed on directories.
type DirOperations interface {
	CreateDirectory(path string) error
	DeleteDirectory(path string) error
	MoveDirectory(sourcePath, destPath string) error
	CopyDirectory(sourcePath, destPath string) error
	ListDirectory(path string) ([]string, error)
	GetDirAttributes(path string) (Directory, error)
}


func (f *FileManagerImpl) CreateDirectory(path string) error {
	config := cm.CommandConfig{
		Command: "mkdir",
		Args:    []string{path},
	}
	_, err := f.commandManager.Run(context.TODO(), config)

	return err
}

func (f *FileManagerImpl) DeleteDirectory(path string) error {
	config := cm.CommandConfig{
		Command: "rm",
		Args:    []string{"-r", path},
	}
	_, err := f.commandManager.Run(context.TODO(), config)
	return err
}

func (f *FileManagerImpl) MoveDirectory(sourcePath, destPath string) error {
	config := cm.CommandConfig{
		Command: "mv",
		Args:    []string{sourcePath, destPath},
	}
	_, err := f.commandManager.Run(context.TODO(), config)
	return err
}

func (f *FileManagerImpl) CopyDirectory(sourcePath, destPath string) error {
	config := cm.CommandConfig{
		Command: "cp",
		Args:    []string{"-r", sourcePath, destPath},
	}
	_, err := f.commandManager.Run(context.TODO(), config)
	return err
}

func (f *FileManagerImpl) ListDirectory(path string) ([]string, error) {
	config := cm.CommandConfig{
		Command: "ls",
		Args:    []string{path},
	}
	result, err := f.commandManager.Run(context.TODO(), config)
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(result.STDOUT), "\n")
	return files, nil
}

func (f *FileManagerImpl) GetDirAttributes(path string) (Directory, error) {
	// For simplicity, we'll only get the modification time and mode.
	// Getting exact attributes like owner and group requires more complex parsing.
	config := cm.CommandConfig{
		Command: "stat",
		Args:    []string{"-c", "%F %Y %a", path}, // Get file type, modification time, and mode
	}
	result, err := f.commandManager.Run(context.TODO(), config)
	if err != nil {
		return Directory{}, err
	}

	// Split the output
	parts := strings.Split(strings.TrimSpace(result.STDOUT), " ")

	if len(parts) != 3 || parts[0] != "directory" {
		return Directory{}, fmt.Errorf("unexpected stat output format or not a directory: %s", result.STDOUT)
	}

	// Extract modification time
	modifiedSeconds, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return Directory{}, fmt.Errorf("error parsing modification time: %v", err)
	}
	modified := time.Unix(modifiedSeconds, 0)

	// Extract mode
	modeInt, err := strconv.ParseInt(parts[2], 8, 64) // octal base
	if err != nil {
		return Directory{}, fmt.Errorf("error parsing mode: %v", err)
	}
	mode := os.FileMode(modeInt)

	return Directory{
		Path:     path,
		Mode:     mode,
		Modified: modified,
	}, nil
}
