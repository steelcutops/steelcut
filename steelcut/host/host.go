package host

import (
	"context"
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
	HostString    string

	PackageManager packagemanager.PackageManager
	NetworkManager networkmanager.NetworkManager
	FileManager    filemanager.FileManager
	HostManager    hostmanager.HostManager
	ServiceManager servicemanager.ServiceManager
	CommandManager commandmanager.CommandManager
}

// DetermineOS method for the ConcreteHost
func (h *ConcreteHost) DetermineOS(ctx context.Context) (OSType, error) {
	cmdConfig := commandmanager.CommandConfig{
		Command: "uname",
		Sudo:    false,
	}

	result, err := h.CommandManager.Run(ctx, h.HostString, cmdConfig)
	if err != nil {
		return Unknown, fmt.Errorf("failed to run uname: %w", err)
	}

	osName := strings.TrimSpace(result.STDOUT)

	switch osName {
	case "Linux":
		return h.detectLinuxType(ctx)
	case "Darwin":
		h.OSType = Darwin
		return Darwin, nil
	default:
		return Unknown, fmt.Errorf("unknown OS: %s", osName)
	}
}

// detectLinuxType method for the ConcreteHost
func (h *ConcreteHost) detectLinuxType(ctx context.Context) (OSType, error) {
	cmdConfig := commandmanager.CommandConfig{
		Command: "cat",
		Args:    []string{"/etc/os-release"},
		Sudo:    false,
	}

	result, err := h.CommandManager.Run(ctx, h.HostString, cmdConfig)
	if err != nil {
		return Unknown, fmt.Errorf("failed to retrieve OS release info: %w", err)
	}

	osRelease := result.STDOUT

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

	return Unknown, fmt.Errorf("unsupported Linux distribution")
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
