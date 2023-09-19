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

type UnixFileManager struct {
	CommandManager *cm.UnixCommandManager
}

func (ufm *UnixFileManager) CreateDirectory(path string) error {
	config := cm.CommandConfig{
		Command: "mkdir",
		Args:    []string{path},
	}
	_, err := ufm.CommandManager.Run(context.TODO(), config)
	return err
}

func (ufm *UnixFileManager) DeleteDirectory(path string) error {
	config := cm.CommandConfig{
		Command: "rm",
		Args:    []string{"-r", path},
	}
	_, err := ufm.CommandManager.Run(context.TODO(), config)
	return err
}

func (ufm *UnixFileManager) MoveDirectory(sourcePath, destPath string) error {
	config := cm.CommandConfig{
		Command: "mv",
		Args:    []string{sourcePath, destPath},
	}
	_, err := ufm.CommandManager.Run(context.TODO(), config)
	return err
}

func (ufm *UnixFileManager) CopyDirectory(sourcePath, destPath string) error {
	config := cm.CommandConfig{
		Command: "cp",
		Args:    []string{"-r", sourcePath, destPath},
	}
	_, err := ufm.CommandManager.Run(context.TODO(), config)
	return err
}

func (ufm *UnixFileManager) ListDirectory(path string) ([]string, error) {
	config := cm.CommandConfig{
		Command: "ls",
		Args:    []string{path},
	}
	result, err := ufm.CommandManager.Run(context.TODO(), config)
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(result.STDOUT), "\n")
	return files, nil
}

func (ufm *UnixFileManager) GetDirAttributes(path string) (Directory, error) {
	// For simplicity, we'll only get the modification time and mode.
	// Getting exact attributes like owner and group requires more complex parsing.
	config := cm.CommandConfig{
		Command: "stat",
		Args:    []string{"-c", "%F %Y %a", path}, // Get file type, modification time, and mode
	}
	result, err := ufm.CommandManager.Run(context.TODO(), config)
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

// FileOperations methods for UnixFileManager

func (ufm *UnixFileManager) CreateFile(path string) error {
	config := cm.CommandConfig{
		Command: "touch",
		Args:    []string{path},
	}
	result, err := ufm.CommandManager.Run(context.TODO(), config)
	return handleCommandResult(result, err)
}

func (ufm *UnixFileManager) DeleteFile(path string) error {
	config := cm.CommandConfig{
		Command: "rm",
		Args:    []string{path},
	}
	result, err := ufm.CommandManager.Run(context.TODO(), config)
	return handleCommandResult(result, err)
}

func (ufm *UnixFileManager) MoveFile(sourcePath, destPath string) error {
	config := cm.CommandConfig{
		Command: "mv",
		Args:    []string{sourcePath, destPath},
	}
	result, err := ufm.CommandManager.Run(context.TODO(), config)
	return handleCommandResult(result, err)
}

func (ufm *UnixFileManager) CopyFile(sourcePath, destPath string) error {
	config := cm.CommandConfig{
		Command: "cp",
		Args:    []string{sourcePath, destPath},
	}
	result, err := ufm.CommandManager.Run(context.TODO(), config)
	return handleCommandResult(result, err)
}

func (ufm *UnixFileManager) GetFileAttributes(path string) (File, error) {
	config := cm.CommandConfig{
		Command: "stat",
		Args:    []string{"-c", "%s %F %Y", path}, // Get size, file type, and modification time
	}
	result, err := ufm.CommandManager.Run(context.TODO(), config)
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

	// Extract FileMode
	var mode os.FileMode
	if parts[1] == "regular file" {
		mode = 0644
	} else if parts[1] == "directory" {
		mode = 0755
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

func (ufm *UnixFileManager) DiskUsage(path string) (DiskUsageInfo, error) {
	config := cm.CommandConfig{
		Command: "df",
		Args:    []string{"-B1", path}, // -B1 ensures output in bytes
	}
	result, err := ufm.CommandManager.Run(context.TODO(), config)
	if err != nil {
		return DiskUsageInfo{}, err
	}

	lines := strings.Split(strings.TrimSpace(result.STDOUT), "\n")
	if len(lines) < 2 {
		return DiskUsageInfo{}, fmt.Errorf("unexpected df output format: %s", result.STDOUT)
	}

	columns := strings.Fields(lines[1])
	if len(columns) < 5 {
		return DiskUsageInfo{}, fmt.Errorf("unexpected df columns count: %s", lines[1])
	}

	total, err := strconv.ParseInt(columns[1], 10, 64)
	if err != nil {
		return DiskUsageInfo{}, fmt.Errorf("error parsing total space: %v", err)
	}

	used, err := strconv.ParseInt(columns[2], 10, 64)
	if err != nil {
		return DiskUsageInfo{}, fmt.Errorf("error parsing used space: %v", err)
	}

	available, err := strconv.ParseInt(columns[3], 10, 64)
	if err != nil {
		return DiskUsageInfo{}, fmt.Errorf("error parsing available space: %v", err)
	}

	usePercentStr := strings.TrimRight(columns[4], "%")
	usePercent, err := strconv.ParseFloat(usePercentStr, 64)
	if err != nil {
		return DiskUsageInfo{}, fmt.Errorf("error parsing usage percentage: %v", err)
	}

	return DiskUsageInfo{
		Total:      total,
		Used:       used,
		Available:  available,
		UsePercent: usePercent,
	}, nil
}
