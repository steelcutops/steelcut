package host

import (
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
