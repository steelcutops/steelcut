package steelcut

import (
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"strings"

	"golang.org/x/crypto/ssh"
)

type SSHClient interface {
	Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error)
}

type RealSSHClient struct{}

func (c RealSSHClient) Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	return ssh.Dial(network, addr, config)
}

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
	UpgradeAllPackages() ([]Update, error)
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

func WithSSHClient(client SSHClient) HostOption {
	return func(h *UnixHost) {
		h.SSHClient = client
	}
}

func WithSudoPassword(password string) HostOption {
	return func(host *UnixHost) {
		host.SudoPassword = password
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
			log.Println("Detected Debian/Ubuntu")
			linuxHost.PackageManager = AptPackageManager{} // Assign PackageManager
		} else {
			// Assume Red Hat/CentOS/Fedora if not Debian/Ubuntu.
			log.Println("Detected Red Hat/CentOS/Fedora")
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

func (h UnixHost) RunCommand(cmd string, options ...interface{}) (string, error) {
	useSudo := false
	sudoPassword := ""

	// Check if options were provided
	if len(options) > 0 {
		if val, ok := options[0].(bool); ok {
			useSudo = val
		}
	}
	if len(options) > 1 {
		if val, ok := options[1].(string); ok {
			sudoPassword = val
		}
	}

	return h.runCommandInternal(cmd, useSudo, sudoPassword)
}

func (h UnixHost) runCommandInternal(cmd string, useSudo bool, sudoPassword string) (string, error) {
	if useSudo {
		log.Printf("Using sudo for command '%s' on host '%s'\n", cmd, h.Hostname)
		cmd = "sudo -S " + cmd // -S option makes sudo read password from standard input
		sudoPassword = h.SudoPassword
	}

	log.Printf("Running command '%s' on host '%s' with user '%s'\n", cmd, h.Hostname, h.User)

	if h.Hostname == "localhost" || h.Hostname == "127.0.0.1" {
		parts := strings.Fields(cmd)
		head := parts[0]
		parts = parts[1:]

		if useSudo && sudoPassword != "" {
			log.Println("Providing sudo password through stdin for local command")
			sudoCmd := append([]string{"-S", head}, parts...)
			command := exec.Command("sudo", sudoCmd...)
			command.Stdin = strings.NewReader(sudoPassword + "\n") // Write password to stdin
			out, err := command.CombinedOutput()
			if err != nil {
				log.Printf("Error running local command with sudo: %v, Output: %s\n", err, string(out))
				return "", err
			}
			return string(out), nil
		} else {
			command := exec.Command(head, parts...)
			out, err := command.Output()
			if err != nil {
				log.Printf("Error running local command: %v\n", err)
				return "", err
			}
			return string(out), nil
		}
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

	client, err := h.SSHClient.Dial("tcp", h.Hostname+":22", config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	if useSudo && sudoPassword != "" {
		session.Stdin = strings.NewReader(sudoPassword + "\n") // Write password to stdin
	}

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		log.Printf("Error running command over SSH with sudo: %v\n", err)
		return "", err
	}

	return string(output), nil
}
