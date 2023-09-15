package host

import (
	"github.com/steelcutops/steelcut/steelcut/packagemanager"
	"github.com/steelcutops/steelcut/steelcut/networkmanager"
	"github.com/steelcutops/steelcut/steelcut/filemanager"
	"github.com/steelcutops/steelcut/steelcut/hostmanager"
	"github.com/steelcutops/steelcut/steelcut/servicemanager"
	"github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type HostInterface interface {
    packagemanager.PackageManager
	networkmanager.NetworkManager
	filemanager.FileManager
	hostmanager.HostManager
	servicemanager.ServiceManager
	commandmanager.CommandManager
}

