package steelcut

import (
	"golang.org/x/crypto/ssh"
	"os/exec"
	"strings"
)

type Update struct {
	// Fields for the Update struct
}

type Host interface {
	CheckUpdates() ([]Update, error)
	RunCommand(cmd string) (string, error)
	DetermineOS() (string, error)
}

type UnixHost struct {
	Hostname string
	User     string
	Password string
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

func NewHost(hostname string, options ...HostOption) (Host, error) {
	host := &UnixHost{
		Hostname: hostname,
	}

	for _, option := range options {
		option(host)
	}

	// Implement the logic to check if the host is reachable or not
	// ...

	return host, nil
}

func (h UnixHost) CheckUpdates() ([]Update, error) {
	// Implement the update check for Unix hosts.
	// The implementation may vary depending on the specific OS.
	return []Update{}, nil
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

func (h UnixHost) DetermineOS() (string, error) {
	// Run the 'uname' command to determine the OS.
	output, err := h.RunCommand("uname")
	if err != nil {
		return "", err
	}

	// The 'uname' command returns a string followed by a newline character.
	// We use strings.TrimSpace to remove the newline character.
	return strings.TrimSpace(output), nil
}
