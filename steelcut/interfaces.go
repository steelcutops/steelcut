package steelcut

import (
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type CommandExecutor interface {
	RunCommand(command string, options CommandOptions) (string, error)
}

// SSHClient defines an interface for dialing and establishing an SSH connection.
type SSHClient interface {
	Dial(network, addr string, config *ssh.ClientConfig, timeout time.Duration) (*ssh.Client, error)
}

type HostInfo struct {
	CPUUsage         float64
	DiskUsage        float64
	MemoryUsage      float64
	RunningProcesses []string
}

// SystemReporter defines an interface for reporting system-related information.
type SystemReporter interface {
	CPUUsage() (float64, error)
	DiskUsage() (float64, error)
	MemoryUsage() (float64, error)
	RunningProcesses() ([]string, error)
	Info() (HostInfo, error)
}

// Host defines an interface for performing operations on a host system.
type Host interface {
	AddPackage(pkg string) error
	CheckUpdates() ([]Update, error)
	Hostname() string
	IsReachable() error
	ListPackages() ([]string, error)
	Reboot() error
	RemovePackage(pkg string) error
	RunCommand(cmd string, options CommandOptions) (string, error)
	Shutdown() error
	SystemReporter
	UpgradeAllPackages() ([]Update, error)
	UpgradePackage(pkg string) error
}

// FileManager defines an interface for performing file management operations.
type FileManager interface {
	CreateDirectory(path string) error
	DeleteDirectory(path string) error
	ListDirectory(path string) ([]string, error)
	SetPermissions(path string, mode os.FileMode) error
	GetPermissions(path string) (os.FileMode, error)
}
