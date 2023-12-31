package host

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/steelcutops/steelcut/common"
	"github.com/steelcutops/steelcut/steelcut/commandmanager"
	"github.com/steelcutops/steelcut/steelcut/filemanager"
	"github.com/steelcutops/steelcut/steelcut/hostmanager"
	"github.com/steelcutops/steelcut/steelcut/networkmanager"
	"github.com/steelcutops/steelcut/steelcut/packagemanager"
	"github.com/steelcutops/steelcut/steelcut/servicemanager"
)

type Host struct {
	common.Credentials

	OSType    OSType
	SSHClient SSHClient
	Hostname  string

	PackageManager packagemanager.PackageManager
	NetworkManager networkmanager.NetworkManager
	FileManager    filemanager.FileManager
	HostManager    hostmanager.HostManager
	ServiceManager servicemanager.ServiceManager
	CommandManager commandmanager.CommandManager
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

// DefaultOSDetector is a default implementation of the OSDetector interface.
type DefaultOSDetector struct{}

// OSType represents various types of Operating Systems that are supported.
type OSType int

const (
	Unknown OSType = iota
	LinuxUbuntu
	LinuxDebian
	LinuxFedora
	LinuxRedHat
	Darwin
	LinuxCentOS
	LinuxArch
	LinuxOpenSUSE
	LinuxAlpine
)

// DetermineOS method for the ConcreteHost
func (h *Host) DetermineOS(ctx context.Context) (OSType, error) {
	cmdConfig := commandmanager.CommandConfig{
		Command: "uname",
		Sudo:    false,
	}

	result, err := h.CommandManager.Run(ctx, cmdConfig)
	slog.Debug("Determine OS result", "hostname", h.Hostname, "result", result, "error", err)
	if err != nil {
		slog.Error("Determine OS error", "hostname", h.Hostname, "error", err)
		return Unknown, fmt.Errorf("failed to run uname: %w", err)
	}
	osName := strings.TrimSpace(result.STDOUT)

	slog.Debug("Determining OS", "hostname", h.Hostname, "osname", osName)

	switch osName {
	case "Linux":
		slog.Debug("Detected Linux", "hostname", h.Hostname)
		return h.detectLinuxType(ctx)
	case "Darwin":
		slog.Debug("Detected Darwin", "hostname", h.Hostname)
		h.OSType = Darwin
		return Darwin, nil
	default:
		slog.Debug("Detected Unknown", "hostname", h.Hostname)
		return Unknown, fmt.Errorf("unknown OS: %s", osName)
	}
}

// detectLinuxType method for the ConcreteHost
func (h *Host) detectLinuxType(ctx context.Context) (OSType, error) {
	cmdConfig := commandmanager.CommandConfig{
		Command: "cat",
		Args:    []string{"/etc/os-release"},
		Sudo:    false,
	}

	result, err := h.CommandManager.Run(ctx, cmdConfig)
	slog.Debug("Detecting Linux type", "hostname", h.Hostname, "result", result, "error", err)
	if err != nil {
		return Unknown, fmt.Errorf("failed to retrieve OS release info: %w", err)
	}

	osRelease := result.STDOUT
	slog.Debug("Detecting Linux type", "hostname", h.Hostname, "osrelease", osRelease)

	if strings.Contains(osRelease, "ID=ubuntu") {
		return LinuxUbuntu, nil
	} else if strings.Contains(osRelease, "ID=debian") {
		return LinuxDebian, nil
	} else if strings.Contains(osRelease, "ID=fedora") {
		return LinuxFedora, nil
	} else if strings.Contains(osRelease, "ID=rhel") {
		return LinuxRedHat, nil
	} else if strings.Contains(osRelease, "ID=centos") {
		return LinuxCentOS, nil
	} else if strings.Contains(osRelease, "ID=arch") {
		return LinuxArch, nil
	} else if strings.Contains(osRelease, "ID=opensuse") {
		return LinuxOpenSUSE, nil
	}

	return Unknown, fmt.Errorf("unsupported Linux distribution detected on host: %s osRelease: %s", h.Hostname, osRelease)
}

// String method provides the string representation of the OSType.
func (o OSType) String() string {
	return [...]string{
		"Unknown",
		"Linux_Ubuntu",
		"Linux_Debian",
		"Linux_Fedora",
		"Linux_RedHat",
		"Darwin",
		"Linux_CentOS",
		"Linux_Arch",
		"Linux_OpenSUSE",
	}[o]
}
