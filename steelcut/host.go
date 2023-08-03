package steelcut

import (
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"strings"

	"golang.org/x/crypto/ssh"
)

type SystemReporter interface {
	CPUUsage() (float64, error)
	MemoryUsage() (float64, error)
	DiskUsage() (float64, error)
	RunningProcesses() ([]string, error)
}

type Host interface {
	CheckUpdates() ([]Update, error)
	RunCommand(cmd string) (string, error)
	ListPackages() ([]string, error)
	AddPackage(pkg string) error
	RemovePackage(pkg string) error
	UpgradePackage(pkg string) error
	Reboot() error
	Shutdown() error
	SystemReporter
}

type HostOption func(*UnixHost)

func WithUser(user string) HostOption {
	return func(host *UnixHost) {
		host.User = user
	}
}

func WithPassword(password string) HostOption {
	return func(host *UnixHost) {
		host.Password = password
	}
}

func WithKeyPassphrase(keyPassphrase string) HostOption {
	return func(host *UnixHost) {
		host.KeyPassphrase = keyPassphrase
	}
}

func WithOS(os string) HostOption {
	return func(host *UnixHost) {
		host.OS = os
	}
}

func determineOS(host *UnixHost) (string, error) {
	output, err := host.RunCommand("uname")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

func NewHost(hostname string, options ...HostOption) (Host, error) {
	unixHost := &UnixHost{
		Hostname: hostname,
	}

	for _, option := range options {
		option(unixHost)
	}

	// If the username has not been specified, use the current user's username.
	if unixHost.User == "" {
		currentUser, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not get current user: %v", err)
		}
		unixHost.User = currentUser.Username
	}

	// If the OS has not been specified, determine it.
	if unixHost.OS == "" {
		os, err := determineOS(unixHost)
		if err != nil {
			return nil, err
		}
		unixHost.OS = os
	}

	switch unixHost.OS {
	case "Linux":
		linuxHost := &LinuxHost{UnixHost: unixHost}
		// Determine the package manager.
		osRelease, _ := linuxHost.RunCommand("cat /etc/os-release")
		if strings.Contains(osRelease, "ID=ubuntu") || strings.Contains(osRelease, "ID=debian") {
			linuxHost.PackageManager = AptPackageManager{} // Assign PackageManager
		} else {
			// Assume Red Hat/CentOS/Fedora if not Debian/Ubuntu.
			linuxHost.PackageManager = YumPackageManager{} // Assign PackageManager
		}
		return linuxHost, nil
	case "Darwin":
		macHost := &MacOSHost{UnixHost: unixHost}
		macHost.PackageManager = BrewPackageManager{} // Assign PackageManager
		return macHost, nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", unixHost.OS)
	}
}

func (h UnixHost) RunCommand(cmd string) (string, error) {
	log.Printf("Running command '%s' on host '%s' with user '%s'\n", cmd, h.Hostname, h.User)
	// If the hostname is "localhost" or "127.0.0.1", run the command locally.
	if h.Hostname == "localhost" || h.Hostname == "127.0.0.1" {
		parts := strings.Fields(cmd)
		head := parts[0]
		parts = parts[1:]

		out, err := exec.Command(head, parts...).Output()
		if err != nil {
			return "", err
		}

		return string(out), nil
	}

	// Otherwise, run the command over SSH.
	var authMethod ssh.AuthMethod

	if h.Password != "" {
		log.Println("Using password authentication")
		authMethod = ssh.Password(h.Password)
	} else {
		log.Println("Using public key authentication")
		var keyManager SSHKeyManager
		if h.KeyPassphrase != "" {
			keyManager = FileSSHKeyManager{}
		} else {
			keyManager = AgentSSHKeyManager{}
		}

		keys, err := keyManager.ReadPrivateKeys(h.KeyPassphrase)
		if err != nil {
			return "", err
		}

		authMethod = ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
			return keys, nil
		})
	}

	config := &ssh.ClientConfig{
		User: h.User,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", h.Hostname+":22", config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
