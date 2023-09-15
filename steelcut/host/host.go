package host

import (
	"fmt"
	"strings"

	"github.com/steelcutops/steelcut/steelcut/commandmanager"
	"github.com/steelcutops/steelcut/steelcut/filemanager"
	"github.com/steelcutops/steelcut/steelcut/hostmanager"
	"github.com/steelcutops/steelcut/steelcut/networkmanager"
	"github.com/steelcutops/steelcut/steelcut/packagemanager"
	"github.com/steelcutops/steelcut/steelcut/servicemanager"
)

type HostInterface interface {
	commandmanager.CommandManager
	filemanager.FileManager
	hostmanager.HostManager
	networkmanager.NetworkManager
	servicemanager.ServiceManager
	packagemanager.PackageManager
}

// ConcreteHost will be the main implementation for both Darwin and Linux.
type ConcreteHost struct {
	User          string
	Password      string
	KeyPassphrase string
	OSType        OSType
	SudoPassword  string
	SSHClient     SSHClient
	HostString    string

	PackageManager packagemanager.PackageManager
	NetworkManager networkmanager.NetworkManager
	FileManager    filemanager.FileManager
	HostManager    hostmanager.HostManager
	ServiceManager servicemanager.ServiceManager
	CommandManager commandmanager.CommandManager
}

// DetermineOS method for the ConcreteHost
func (h *ConcreteHost) DetermineOS() (OSType, error) {
	output, err := h.CommandManager.Run("uname", false)
	if err != nil {
		return Unknown, fmt.Errorf("failed to run uname: %w", err)
	}

	osName := strings.TrimSpace(output)

	switch osName {
	case "Linux":
		return h.detectLinuxType()
	case "Darwin":
		h.OSType = Darwin
		return Darwin, nil
	default:
		return Unknown, fmt.Errorf("unknown OS: %s", osName)
	}
}

// detectLinuxType method for the ConcreteHost
func (h *ConcreteHost) detectLinuxType() (OSType, error) {
	osRelease, err := h.CommandManager.Run("cat /etc/os-release", false)
	if err != nil {
		return Unknown, fmt.Errorf("failed to retrieve OS release info: %w", err)
	}

	// ... (rest of the logic remains unchanged)
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
)

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