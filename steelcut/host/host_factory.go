package host

import (
	"fmt"
	"github.com/steelcutops/steelcut/steelcut/commandmanager"
	"github.com/steelcutops/steelcut/steelcut/filemanager"
)

func NewHost() (HostInterface, error) {
	var ch ConcreteHost
	osType, err := ch.DetermineOS()
	if err != nil {
		return nil, err
	}

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
		FileManager:    filemanager.NewFileManager("localhost", &commandmanager.UnixCommandManager{}),
		HostManager:    &LinuxHostManager{},
		NetworkManager: &LinuxNetworkManager{},
		ServiceManager: &LinuxServiceManager{},
		PackageManager: &LinuxPackageManager{},
	}
}

func configureMacHost() ConcreteHost {
	return ConcreteHost{
		CommandManager: &commandmanager.UnixCommandManager{},
		FileManager:    filemanager.NewFileManager("localhost", &commandmanager.UnixCommandManager{}),
		HostManager:    &DarwinHostManager{},
		NetworkManager: &DarwinNetworkManager{},
		ServiceManager: &DarwinServiceManager{},
		PackageManager: &DarwinPackageManager{},
	}
}
