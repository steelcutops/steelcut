package host

import (
	"context"
	"fmt"

	"github.com/steelcutops/steelcut/steelcut/commandmanager"
	"github.com/steelcutops/steelcut/steelcut/filemanager"
	"github.com/steelcutops/steelcut/steelcut/hostmanager"
	"github.com/steelcutops/steelcut/steelcut/networkmanager"
	"github.com/steelcutops/steelcut/steelcut/packagemanager"
	"github.com/steelcutops/steelcut/steelcut/servicemanager"
)

func NewHost(hostname string, options ...HostOption) (*Host, error) {
    ch := &Host{}

    // Apply each HostOption
    for _, option := range options {
        option(ch)
    }

    // Initializing the CommandManager is required before determining the OS
    ch.CommandManager = &commandmanager.UnixCommandManager{Hostname: hostname}

    osType, err := ch.DetermineOS(context.TODO())
    if err != nil {
        return nil, err
    }

    switch osType {
    case LinuxUbuntu, LinuxDebian, LinuxFedora, LinuxRedHat, LinuxCentOS, LinuxArch, LinuxOpenSUSE:
        configureLinuxHost(ch, hostname, osType)
    case Darwin:
        configureMacHost(ch, hostname)
    default:
        return nil, fmt.Errorf("unsupported operating system: %s", osType)
    }

    return ch, nil
}


func configureLinuxHost(ch *Host, hostname string, osType OSType) {
	cmdManager := &commandmanager.UnixCommandManager{Hostname: hostname}
	var pkgManager packagemanager.PackageManager

	switch osType {
	case LinuxUbuntu, LinuxDebian:
		pkgManager = &packagemanager.AptPackageManager{CommandManager: cmdManager}
	case LinuxFedora:
		pkgManager = &packagemanager.DnfPackageManager{CommandManager: cmdManager}
	case LinuxRedHat, LinuxCentOS:
		pkgManager = &packagemanager.YumPackageManager{CommandManager: cmdManager}
	default:
		pkgManager = nil
	}

	ch.CommandManager = cmdManager
	ch.FileManager = &filemanager.UnixFileManager{CommandManager: cmdManager}
	ch.HostManager = &hostmanager.UnixHostManager{CommandManager: cmdManager}
	ch.NetworkManager = &networkmanager.UnixNetworkManager{CommandManager: cmdManager}
	ch.ServiceManager = &servicemanager.LinuxServiceManager{CommandManager: cmdManager}
	ch.PackageManager = pkgManager
}

func configureMacHost(ch *Host, hostname string) {
	cmdManager := &commandmanager.UnixCommandManager{Hostname: hostname}

	ch.CommandManager = cmdManager
	ch.FileManager = &filemanager.UnixFileManager{CommandManager: cmdManager}
	ch.HostManager = &hostmanager.UnixHostManager{CommandManager: cmdManager}
	ch.NetworkManager = &networkmanager.UnixNetworkManager{CommandManager: cmdManager}
	ch.ServiceManager = &servicemanager.DarwinServiceManager{CommandManager: cmdManager}
	ch.PackageManager = &packagemanager.BrewPackageManager{CommandManager: cmdManager}
}
