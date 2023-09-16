package host

import (
	"fmt"
	"github.com/steelcutops/steelcut/steelcut/commandmanager"
)

func NewHost() (HostInterface, error) {
	osType, err := DetermineOS()
	if err != nil {
		return nil, err
	}

	var ch ConcreteHost

	switch osType {
	case LinuxUbuntu, LinuxDebian, LinuxFedora, LinuxRedHat, LinuxCentOS, LinuxArch, LinuxOpenSUSE:
		ch = configureLinuxHost()
	case Darwin:
		ch = configureMacHost()
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", osType)
	}

	return &ch, nil
}

func configureLinuxHost() ConcreteHost {
	return ConcreteHost{
		CommandManager: &commandmanager.UnixCommandManager{},
		FileManager:    &LinuxFileManager{},
		HostManager:    &LinuxHostManager{},
		NetworkManager: &LinuxNetworkManager{},
		ServiceManager: &LinuxServiceManager{},
		PackageManager: &LinuxPackageManager{},
	}
}

func configureMacHost() ConcreteHost {
	return ConcreteHost{
		CommandManager: &commandmanager.UnixCommandManager{},
		FileManager:    &DarwinFileManager{},
		HostManager:    &DarwinHostManager{},
		NetworkManager: &DarwinNetworkManager{},
		ServiceManager: &DarwinServiceManager{},
		PackageManager: &DarwinPackageManager{},
	}
}
