package steelcut

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// UnixHost represents a Unix-based host system with details like username, password, and connection information.
type UnixHost struct {
	User          string
	Password      string
	KeyPassphrase string
	OS            string
	SudoPassword  string
	SSHClient     SSHClient
	HostString    string
	FileManager   FileManager
	Executor      CommandExecutor
	PkgManager    PackageManager
}

// Hostname returns the host string (e.g., IP or domain name) for the Unix host.
func (h *UnixHost) Hostname() string {
	return h.HostString // Return the renamed field
}

// CreateDirectory creates a directory at the given path.
func (h UnixHost) CreateDirectory(path string) error {
	_, err := h.runCommandInternal(fmt.Sprintf("mkdir -p %s", path), false, "")
	return err
}

// DeleteDirectory deletes the directory at the given path, including all contents.
func (h UnixHost) DeleteDirectory(path string) error {
	_, err := h.runCommandInternal(fmt.Sprintf("rm -rf %s", path), false, "")
	return err
}

// ListDirectory lists the contents of the directory at the given path.
func (h UnixHost) ListDirectory(path string) ([]string, error) {
	output, err := h.runCommandInternal(fmt.Sprintf("ls %s", path), false, "")
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(output), "\n"), nil
}

// SetPermissions sets the file permissions for the given path, using the provided FileMode.
func (h UnixHost) SetPermissions(path string, mode os.FileMode) error {
	_, err := h.runCommandInternal(fmt.Sprintf("chmod %o %s", mode, path), false, "")
	return err
}

// GetPermissions retrieves the file permissions for the given path, returning them as a FileMode value.
func (h UnixHost) GetPermissions(path string) (os.FileMode, error) {
	output, err := h.runCommandInternal(fmt.Sprintf("stat -c %%a %s", path), false, "")
	if err != nil {
		return 0, err
	}
	mode, err := strconv.ParseUint(strings.TrimSpace(output), 8, 32)
	if err != nil {
		return 0, err
	}
	return os.FileMode(mode), nil
}
