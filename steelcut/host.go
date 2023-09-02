// Package steelcut provides functionalities to manage Unix hosts, perform SSH connections,
// report system-related information, and manage files and directories.
package steelcut

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type CommandExecutor interface {
	RunCommand(command string, options CommandOptions) (string, error)
}

type commandResult struct {
	output []byte
	err    error
}

type OSDetector interface {
	DetermineOS(host *UnixHost) (string, error)
}

type DefaultOSDetector struct{}

func (d DefaultOSDetector) DetermineOS(host *UnixHost) (string, error) {
	output, err := host.RunCommand("uname", CommandOptions{
		UseSudo: false,
	})
	if err != nil {
		return "", err
	}

	osType := strings.TrimSpace(output)

	if osType == "Linux" {
		osRelease, err := host.RunCommand("cat /etc/os-release", CommandOptions{UseSudo: false})
		if err != nil {
			return "", err
		}

		if strings.Contains(osRelease, "ID=ubuntu") || strings.Contains(osRelease, "ID=debian") {
			return "Linux_Ubuntu", nil
		} else if strings.Contains(osRelease, "ID=fedora") {
			return "Linux_Fedora", nil
		} else {
			return "Linux_RedHat", nil
		}
	}

	return osType, nil
}

// SSHClient defines an interface for dialing and establishing an SSH connection.
type SSHClient interface {
	Dial(network, addr string, config *ssh.ClientConfig, timeout time.Duration) (*ssh.Client, error)
}

// RealSSHClient provides a real implementation of the SSHClient interface.
type RealSSHClient struct{}

// Dial dials an SSH connection with the given network, address, client config, and timeout.
func (c RealSSHClient) Dial(network, addr string, config *ssh.ClientConfig, timeout time.Duration) (*ssh.Client, error) {
	// Dial with a timeout
	conn, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}

	// Create an SSH client connection using the underlying network connection
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(sshConn, chans, reqs), nil
}

type DefaultCommandExecutor struct {
	Host    Host
	Options CommandOptions
}

func (dce DefaultCommandExecutor) RunCommand(command string, options CommandOptions) (string, error) {
	if options == (CommandOptions{}) { // if no specific options provided
		options = dce.Options // use the defaults
	}

	return dce.RunCommandWithOverride(command, options)
}

func (dce DefaultCommandExecutor) RunCommandWithOverride(command string, overrideOptions CommandOptions) (string, error) {
	if dce.Host == nil {
		return "", errors.New("host is not set in command executor")
	}

	finalOptions := dce.Options // Start with default options.

	// Override with provided options if necessary.
	if overrideOptions.UseSudo {
		finalOptions.UseSudo = overrideOptions.UseSudo
	}
	if overrideOptions.SudoPassword != "" {
		finalOptions.SudoPassword = overrideOptions.SudoPassword
	}

	return dce.Host.RunCommand(command, finalOptions)
}

type CommandOptions struct {
	UseSudo      bool
	SudoPassword string
}

// SystemReporter defines an interface for reporting system-related information.
type SystemReporter interface {
	CPUUsage() (float64, error)
	DiskUsage() (float64, error)
	MemoryUsage() (float64, error)
	RunningProcesses() ([]string, error)
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

type HostOption func(*UnixHost)

// WithUser returns a HostOption that sets the user for a UnixHost.
func WithUser(user string) HostOption {
	return func(host *UnixHost) {
		host.User = user
	}
}

// WithPassword returns a HostOption that sets the password for a UnixHost.
func WithPassword(password string) HostOption {
	return func(host *UnixHost) {
		host.Password = password
	}
}

// WithKeyPassphrase returns a HostOption that sets the key passphrase for a UnixHost.
func WithKeyPassphrase(keyPassphrase string) HostOption {
	return func(host *UnixHost) {
		host.KeyPassphrase = keyPassphrase
	}
}

// WithOS returns a HostOption that sets the OS for a UnixHost.
func WithOS(os string) HostOption {
	return func(host *UnixHost) {
		host.OS = os
	}
}

// WithSSHClient returns a HostOption that sets the SSHClient for a UnixHost.
func WithSSHClient(client SSHClient) HostOption {
	return func(h *UnixHost) {
		h.SSHClient = client
	}
}

// WithSudoPassword returns a HostOption that sets the sudo password for a UnixHost.
func WithSudoPassword(password string) HostOption {
	return func(host *UnixHost) {
		host.SudoPassword = password
	}
}

func WithCommandExecutor(executor CommandExecutor) HostOption {
	return func(h *UnixHost) {
		h.Executor = executor
	}
}

func WithOSDetector(detector OSDetector) HostOption {
	return func(host *UnixHost) {
		host.Detector = detector
	}
}

func NewHost(hostname string, options ...HostOption) (Host, error) {
	unixHost := &UnixHost{
		HostString: hostname,
	}

	for _, option := range options {
		option(unixHost)
	}

	if err := setDefaultUserIfEmpty(unixHost); err != nil {
		return nil, err
	}

	if unixHost.Detector == nil {
		unixHost.Detector = DefaultOSDetector{}
	}

	// If the OS has not been specified, determine it.
	if unixHost.OS == "" {
		osType, err := unixHost.Detector.DetermineOS(unixHost)
		if err != nil {
			return nil, err
		}
		unixHost.OS = osType
	}

	cmdOptions := CommandOptions{
		SudoPassword: unixHost.SudoPassword,
	}

	switch {
	case isOsType(unixHost.OS, "Linux_Ubuntu", "Linux_Debian"):
		return configureLinuxHost(unixHost, cmdOptions, "apt"), nil
	case isOsType(unixHost.OS, "Linux_RedHat", "Linux_CentOS"):
		return configureLinuxHost(unixHost, cmdOptions, "yum"), nil
	case isOsType(unixHost.OS, "Linux_Fedora"):
		return configureLinuxHost(unixHost, cmdOptions, "dnf"), nil
	case unixHost.OS == "Darwin":
		return configureMacHost(unixHost, cmdOptions), nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", unixHost.OS)
	}

}

func setDefaultUserIfEmpty(host *UnixHost) error {
	if host.User != "" {
		return nil
	}
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user: %v", err)
	}
	host.User = currentUser.Username
	return nil
}

func isOsType(os string, types ...string) bool {
	for _, t := range types {
		if strings.HasPrefix(os, t) {
			return true
		}
	}
	return false
}

func configureLinuxHost(host *UnixHost, cmdOptions CommandOptions, pkgManagerType string) *LinuxHost {
	linuxHost := &LinuxHost{UnixHost: host}
	if host.Executor == nil {
		host.Executor = &DefaultCommandExecutor{
			Host:    linuxHost,
			Options: cmdOptions,
		}
	}

	switch pkgManagerType {
	case "apt":
		linuxHost.PackageManager = AptPackageManager{Executor: host.Executor}
	case "yum":
		linuxHost.PackageManager = YumPackageManager{Executor: host.Executor}
	case "dnf":
		linuxHost.PackageManager = DnfPackageManager{Executor: host.Executor}
	}

	return linuxHost
}

func configureMacHost(host *UnixHost, cmdOptions CommandOptions) *MacOSHost {
	macHost := &MacOSHost{UnixHost: host}
	if host.Executor == nil {
		host.Executor = &DefaultCommandExecutor{
			Host:    macHost,
			Options: cmdOptions,
		}
	}
	macHost.PackageManager = BrewPackageManager{Executor: host.Executor}
	return macHost
}
