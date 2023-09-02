// Package steelcut provides functionalities to manage Unix hosts, perform SSH connections,
// report system-related information, and manage files and directories.
package steelcut

import (
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type commandResult struct {
	output []byte
	err    error
}

type OSDetector interface {
	DetermineOS(host *UnixHost) (string, error)
}

type DefaultOSDetector struct{}

func (d DefaultOSDetector) DetermineOS(host *UnixHost) (string, error) {
	output, err := host.RunCommand("uname", CommandOptions{
		UseSudo: false,
	})
	if err != nil {
		return "", err
	}

	osType := strings.TrimSpace(output)

	if osType == "Linux" {
		osRelease, err := host.RunCommand("cat /etc/os-release", CommandOptions{UseSudo: false})
		if err != nil {
			return "", err
		}

		if strings.Contains(osRelease, "ID=ubuntu") || strings.Contains(osRelease, "ID=debian") {
			return "Linux_Ubuntu", nil
		} else if strings.Contains(osRelease, "ID=fedora") {
			return "Linux_Fedora", nil
		} else {
			return "Linux_RedHat", nil
		}
	}

	return osType, nil
}

// RealSSHClient provides a real implementation of the SSHClient interface.
type RealSSHClient struct{}

// Dial dials an SSH connection with the given network, address, client config, and timeout.
func (c RealSSHClient) Dial(network, addr string, config *ssh.ClientConfig, timeout time.Duration) (*ssh.Client, error) {
	// Dial with a timeout
	conn, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}

	// Create an SSH client connection using the underlying network connection
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(sshConn, chans, reqs), nil
}

type CommandOptions struct {
	UseSudo      bool
	SudoPassword string
}

func NewHost(hostname string, options ...HostOption) (Host, error) {
	unixHost := &UnixHost{
		HostString: hostname,
	}

	for _, option := range options {
		option(unixHost)
	}

	if err := setDefaultUserIfEmpty(unixHost); err != nil {
		return nil, err
	}

	if unixHost.Detector == nil {
		unixHost.Detector = DefaultOSDetector{}
	}

	// If the OS has not been specified, determine it.
	if unixHost.OS == "" {
		osType, err := unixHost.Detector.DetermineOS(unixHost)
		if err != nil {
			return nil, err
		}
		unixHost.OS = osType
	}

	cmdOptions := CommandOptions{
		SudoPassword: unixHost.SudoPassword,
	}

	switch {
	case isOsType(unixHost.OS, "Linux_Ubuntu", "Linux_Debian"):
		return configureLinuxHost(unixHost, cmdOptions, "apt"), nil
	case isOsType(unixHost.OS, "Linux_RedHat", "Linux_CentOS"):
		return configureLinuxHost(unixHost, cmdOptions, "yum"), nil
	case isOsType(unixHost.OS, "Linux_Fedora"):
		return configureLinuxHost(unixHost, cmdOptions, "dnf"), nil
	case unixHost.OS == "Darwin":
		return configureMacHost(unixHost, cmdOptions), nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", unixHost.OS)
	}

}

func configureLinuxHost(host *UnixHost, cmdOptions CommandOptions, pkgManagerType string) *LinuxHost {
	linuxHost := &LinuxHost{UnixHost: host}
	if host.Executor == nil {
		host.Executor = &DefaultCommandExecutor{
			Host:    linuxHost,
			Options: cmdOptions,
		}
	}

	switch pkgManagerType {
	case "apt":
		linuxHost.PackageManager = AptPackageManager{Executor: host.Executor}
	case "yum":
		linuxHost.PackageManager = YumPackageManager{Executor: host.Executor}
	case "dnf":
		linuxHost.PackageManager = DnfPackageManager{Executor: host.Executor}
	}

	return linuxHost
}

func configureMacHost(host *UnixHost, cmdOptions CommandOptions) *MacOSHost {
	macHost := &MacOSHost{UnixHost: host}
	if host.Executor == nil {
		host.Executor = &DefaultCommandExecutor{
			Host:    macHost,
			Options: cmdOptions,
		}
	}
	macHost.PackageManager = BrewPackageManager{Executor: host.Executor}
	return macHost
}
