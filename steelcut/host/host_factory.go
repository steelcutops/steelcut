package host

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/user"

	"github.com/steelcutops/steelcut/steelcut/commandmanager"
	"github.com/steelcutops/steelcut/steelcut/filemanager"
	"github.com/steelcutops/steelcut/steelcut/hostmanager"
	"github.com/steelcutops/steelcut/steelcut/networkmanager"
	"github.com/steelcutops/steelcut/steelcut/packagemanager"
	"github.com/steelcutops/steelcut/steelcut/servicemanager"
)

func NewHost(hostname string, options ...HostOption) (*Host, error) {
	ch := &Host{}
	ch.Hostname = hostname

	// Apply each HostOption
	for _, option := range options {
		option(ch)
	}

	// If SSHClient hasn't been set, set it to the default SSHClient
	if ch.SSHClient == nil {
		slog.Debug("SSHClient is nil, setting to default SSHClient")
		ch.SSHClient = &RealSSHClient{}
	} else {
		slog.Debug("SSHClient is not nil, using provided SSHClient", "sshclient", ch.SSHClient)
	}

	// If User hasn't been set, set it to the username of the current user
	if ch.Credentials.User == "" {
		currentUser, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not get current user: %v", err)
		}
		ch.Credentials.User = currentUser.Username
	}

	// If SudoPassword hasn't been set, check environment variables for it
	if ch.Credentials.SudoPassword == "" {
		steelcutBecomePass := os.Getenv("STEELCUT_BECOME_PASS")
		if steelcutBecomePass != "" {
			ch.Credentials.SudoPassword = steelcutBecomePass
		} else {
			ansibleBecomePass := os.Getenv("ANSIBLE_BECOME_PASS")
			if ansibleBecomePass != "" {
				ch.Credentials.SudoPassword = ansibleBecomePass
			}
		}
	}

	// Initializing the CommandManager with the new interface
	ch.CommandManager = &commandmanager.UnixCommandManager{
		Hostname:    hostname,
		Credentials: ch.Credentials,
		SSHClient:   ch.SSHClient,
	}

	osType, err := ch.DetermineOS(context.TODO())
	if err != nil {
		return nil, err
	}

	switch osType {
	case LinuxUbuntu, LinuxDebian, LinuxFedora, LinuxRedHat, LinuxCentOS, LinuxArch, LinuxOpenSUSE:
		configureLinuxHost(ch, ch.CommandManager, osType)

	case Darwin:
		configureMacHost(ch, ch.CommandManager)
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", osType)
	}

	return ch, nil
}

func configureLinuxHost(ch *Host, cmdManager commandmanager.CommandManager, osType OSType) {
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

func configureMacHost(ch *Host, cmdManager commandmanager.CommandManager) {
	ch.CommandManager = cmdManager
	ch.FileManager = &filemanager.UnixFileManager{CommandManager: cmdManager}
	ch.HostManager = &hostmanager.UnixHostManager{CommandManager: cmdManager}
	ch.NetworkManager = &networkmanager.UnixNetworkManager{CommandManager: cmdManager}
	ch.ServiceManager = &servicemanager.DarwinServiceManager{CommandManager: cmdManager}
	ch.PackageManager = &packagemanager.BrewPackageManager{CommandManager: cmdManager}
}
