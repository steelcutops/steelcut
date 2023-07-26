package steelcut

import (
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Update struct {
	// Fields for the Update struct
}

type Host interface {
	CheckUpdates() ([]Update, error)
	RunCommand(cmd string) (string, error)
}

type UnixHost struct {
	Hostname string
	User     string
	Password string
	OS       string
}

type LinuxHost struct {
	UnixHost
}

type MacOSHost struct {
	UnixHost
}

func (h LinuxHost) CheckUpdates() ([]Update, error) {
	// Implement the update check for Linux hosts.
	return []Update{}, nil
}

func (h LinuxHost) RunCommand(cmd string) (string, error) {
	return h.UnixHost.RunCommand(cmd)
}

func (h MacOSHost) CheckUpdates() ([]Update, error) {
	// Implement the update check for macOS hosts.
	return []Update{}, nil
}

func (h MacOSHost) RunCommand(cmd string) (string, error) {
	return h.UnixHost.RunCommand(cmd)
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

func WithOS(os string) HostOption {
	return func(host *UnixHost) {
		host.OS = os
	}
}

func NewHost(hostname string, options ...HostOption) (Host, error) {
	host := &UnixHost{
		Hostname: hostname,
	}

	for _, option := range options {
		option(host)
	}

	// If the OS has not been specified, determine it.
	if host.OS == "" {
		os, err := determineOS(host)
		if err != nil {
			return nil, err
		}
		host.OS = os
	}

	// Return a different type of Host based on the OS.
	switch host.OS {
	case "Linux":
		return LinuxHost{*host}, nil
	case "Darwin":
		return MacOSHost{*host}, nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", host.OS)
	}
}

func determineOS(host *UnixHost) (string, error) {
	output, err := host.RunCommand("uname")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

func (h UnixHost) RunCommand(cmd string) (string, error) {
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
	config := &ssh.ClientConfig{
		User: h.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(h.Password),
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
