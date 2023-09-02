// Package steelcut provides functionalities to manage Unix hosts,
// perform SSH connections, report system-related information,
// and manage files and directories.
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

// OSDetector is an interface that provides a method to determine the OS type of a Unix host.
type OSDetector interface {
	DetermineOS(host *UnixHost) (OSType, error)
}

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
	}[o]
}

// DetermineOS modifies the original function to return an OSType enum.
func (d DefaultOSDetector) DetermineOS(host *UnixHost) (OSType, error) {
	output, err := host.RunCommand("uname", CommandOptions{
		UseSudo: false,
	})
	if err != nil {
		return Unknown, err
	}

	osName := strings.TrimSpace(output)

	if osName == "Linux" {
		osRelease, err := host.RunCommand("cat /etc/os-release", CommandOptions{UseSudo: false})
		if err != nil {
			return Unknown, err
		}

		if strings.Contains(osRelease, "ID=ubuntu") || strings.Contains(osRelease, "ID=debian") {
			return LinuxUbuntu, nil
		} else if strings.Contains(osRelease, "ID=fedora") {
			return LinuxFedora, nil
		} else {
			return LinuxRedHat, nil
		}
	} else if osName == "Darwin" {
		return Darwin, nil
	}

	return Unknown, nil
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

// CommandOptions struct holds options for running a command on a Unix host.
type CommandOptions struct {
	UseSudo      bool
	SudoPassword string
}

// NewHost initializes a new UnixHost based on the provided options and performs OS detection.
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
	if unixHost.OSType == Unknown { // Assuming that OSType field is added to UnixHost
		osType, err := unixHost.Detector.DetermineOS(unixHost)
		if err != nil {
			return nil, err
		}
		unixHost.OSType = osType
	}

	cmdOptions := CommandOptions{
		SudoPassword: unixHost.SudoPassword,
	}

	switch unixHost.OSType { // Updated to use OSType enum
	case LinuxUbuntu, LinuxDebian:
		return configureLinuxHost(unixHost, cmdOptions, "apt"), nil
	case LinuxRedHat:
		return configureLinuxHost(unixHost, cmdOptions, "yum"), nil
	case LinuxFedora:
		return configureLinuxHost(unixHost, cmdOptions, "dnf"), nil
	case Darwin:
		return configureMacHost(unixHost, cmdOptions), nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", unixHost.OSType)
	}
}

// configureLinuxHost configures a Linux host given a Unix host and package manager type.
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

// configureMacHost configures a macOS host given a Unix host.
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
