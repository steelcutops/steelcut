package steelcut

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type UnixHost struct {
	User          string
	Password      string
	KeyPassphrase string
	OS            string
	SudoPassword  string
	SSHClient     SSHClient
	HostString    string
	FileManager   FileManager
}

func (h *UnixHost) Hostname() string {
	return h.HostString // Return the renamed field
}

func (h UnixHost) CreateDirectory(path string) error {
	_, err := h.runCommandInternal(fmt.Sprintf("mkdir -p %s", path), false, "")
	return err
}

func (h UnixHost) DeleteDirectory(path string) error {
	_, err := h.runCommandInternal(fmt.Sprintf("rm -rf %s", path), false, "")
	return err
}

func (h UnixHost) ListDirectory(path string) ([]string, error) {
	output, err := h.runCommandInternal(fmt.Sprintf("ls %s", path), false, "")
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(output), "\n"), nil
}

func (h UnixHost) SetPermissions(path string, mode os.FileMode) error {
	_, err := h.runCommandInternal(fmt.Sprintf("chmod %o %s", mode, path), false, "")
	return err
}

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
