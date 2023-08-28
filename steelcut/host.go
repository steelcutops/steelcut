// Package steelcut provides functionalities to manage Unix hosts, perform SSH connections,
// report system-related information, and manage files and directories.
package steelcut

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
)

type CommandExecutor interface {
	RunCommand(command string, options CommandOptions) (string, error)
}

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

// SSHClient defines an interface for dialing and establishing an SSH connection.
type SSHClient interface {
	Dial(network, addr string, config *ssh.ClientConfig, timeout time.Duration) (*ssh.Client, error)
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

type DefaultCommandExecutor struct {
	Host    Host
	Options CommandOptions
}

func (dce DefaultCommandExecutor) RunCommand(command string, options CommandOptions) (string, error) {
	if options == (CommandOptions{}) { // if no specific options provided
		options = dce.Options // use the defaults
	}

	return dce.RunCommandWithOverride(command, options)
}

func (dce DefaultCommandExecutor) RunCommandWithOverride(command string, overrideOptions CommandOptions) (string, error) {
	if dce.Host == nil {
		return "", errors.New("host is not set in command executor")
	}

	finalOptions := dce.Options // Start with default options.

	// Override with provided options if necessary.
	if overrideOptions.UseSudo {
		finalOptions.UseSudo = overrideOptions.UseSudo
	}
	if overrideOptions.SudoPassword != "" {
		finalOptions.SudoPassword = overrideOptions.SudoPassword
	}

	return dce.Host.RunCommand(command, finalOptions)
}

type CommandOptions struct {
	UseSudo      bool
	SudoPassword string
}

// SystemReporter defines an interface for reporting system-related information.
type SystemReporter interface {
	CPUUsage() (float64, error)
	DiskUsage() (float64, error)
	MemoryUsage() (float64, error)
	RunningProcesses() ([]string, error)
}

// Host defines an interface for performing operations on a host system.
type Host interface {
	AddPackage(pkg string) error
	CheckUpdates() ([]Update, error)
	Hostname() string
	IsReachable() error
	ListPackages() ([]string, error)
	Reboot() error
	RemovePackage(pkg string) error
	RunCommand(cmd string, options CommandOptions) (string, error)
	Shutdown() error
	SystemReporter
	UpgradeAllPackages() ([]Update, error)
	UpgradePackage(pkg string) error
}

// FileManager defines an interface for performing file management operations.
type FileManager interface {
	CreateDirectory(path string) error
	DeleteDirectory(path string) error
	ListDirectory(path string) ([]string, error)
	SetPermissions(path string, mode os.FileMode) error
	GetPermissions(path string) (os.FileMode, error)
}

type HostOption func(*UnixHost)

// WithUser returns a HostOption that sets the user for a UnixHost.
func WithUser(user string) HostOption {
	return func(host *UnixHost) {
		host.User = user
	}
}

// WithPassword returns a HostOption that sets the password for a UnixHost.
func WithPassword(password string) HostOption {
	return func(host *UnixHost) {
		host.Password = password
	}
}

// WithKeyPassphrase returns a HostOption that sets the key passphrase for a UnixHost.
func WithKeyPassphrase(keyPassphrase string) HostOption {
	return func(host *UnixHost) {
		host.KeyPassphrase = keyPassphrase
	}
}

// WithOS returns a HostOption that sets the OS for a UnixHost.
func WithOS(os string) HostOption {
	return func(host *UnixHost) {
		host.OS = os
	}
}

// WithSSHClient returns a HostOption that sets the SSHClient for a UnixHost.
func WithSSHClient(client SSHClient) HostOption {
	return func(h *UnixHost) {
		h.SSHClient = client
	}
}

// WithSudoPassword returns a HostOption that sets the sudo password for a UnixHost.
func WithSudoPassword(password string) HostOption {
	return func(host *UnixHost) {
		host.SudoPassword = password
	}
}

func WithCommandExecutor(executor CommandExecutor) HostOption {
	return func(h *UnixHost) {
		h.Executor = executor
	}
}

func WithOSDetector(detector OSDetector) HostOption {
	return func(host *UnixHost) {
		host.Detector = detector
	}
}

func (h UnixHost) IsReachable() error {
	if h.isLocal() {
		return nil
	}

	if err := h.ping(); err != nil {
		return err
	}
	return h.sshable()
}

func (h UnixHost) ping() error {
	cmd := "ping -c 1 " + h.Hostname()
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return fmt.Errorf("ping test failed: %v", err)
	}
	log.Printf("Ping test passed for host '%s'\n", h.Hostname())
	return nil
}

func (h UnixHost) sshable() error {
	if h.isLocal() {
		return nil
	}

	config, err := h.getSSHConfig()
	if err != nil {
		return err
	}

	timeout := 60 * time.Second
	client, err := h.SSHClient.Dial("tcp", h.Hostname()+":22", config, timeout)
	if err != nil {
		return fmt.Errorf("SSH test failed: %v", err)
	}
	client.Close()
	log.Printf("SSH test passed for host '%s'\n", h.Hostname())
	return nil
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

func setDefaultUserIfEmpty(host *UnixHost) error {
	if host.User != "" {
		return nil
	}
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user: %v", err)
	}
	host.User = currentUser.Username
	return nil
}

func isOsType(os string, types ...string) bool {
	for _, t := range types {
		if strings.HasPrefix(os, t) {
			return true
		}
	}
	return false
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

// RunCommand executes the specified command on the host, either locally or remotely via SSH.
// It takes the command string to be executed and optional parameters to modify the execution.
// Supported options include using sudo for superuser privileges and providing a sudo password.
// Returns the output of the command and an error if an error occurs during execution.
func (h UnixHost) RunCommand(cmd string, options CommandOptions) (string, error) {
	return h.runCommandInternal(cmd, options.UseSudo, options.SudoPassword)
}

// CopyFile copies a file from the local path to the remote path on the host.
func (h UnixHost) CopyFile(localPath string, remotePath string) error {
	// Check if the operation is local
	if h.isLocal() {
		return errors.New("source and destination are the same host")
	}

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	// Get file stats
	fileInfo, err := localFile.Stat()
	if err != nil {
		return err
	}

	// Get SSH client config
	config, err := h.getSSHConfig()
	if err != nil {
		return err
	}

	// Dial SSH connection
	timeout := 30 * time.Second
	client, err := h.SSHClient.Dial("tcp", h.Hostname()+":22", config, timeout)
	if err != nil {
		return err
	}
	defer client.Close()

	// Start a new session
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Start SCP in the remote machine
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintln(w, "C0644", fileInfo.Size(), filepath.Base(remotePath))
		io.Copy(w, localFile)
		fmt.Fprint(w, "\x00")
	}()

	// Run SCP on the remote machine to copy the file
	cmd := "scp -t " + remotePath
	if err := session.Run(cmd); err != nil {
		return err
	}

	log.Printf("File copied successfully from '%s' to '%s'\n", localPath, remotePath)
	return nil
}

func (h UnixHost) runCommandInternal(cmd string, useSudo bool, sudoPassword string) (string, error) {
	if h.isLocal() {
		return h.runLocalCommand(cmd, useSudo, sudoPassword)
	}

	return h.runRemoteCommand(cmd, useSudo, sudoPassword)
}

func (h UnixHost) isLocal() bool {
	return h.Hostname() == "localhost" || h.Hostname() == "127.0.0.1"
}

func (h UnixHost) runLocalCommand(cmd string, useSudo bool, sudoPassword string) (string, error) {
	if useSudo {
		if sudoPassword == "" {
			return "", errors.New("sudo: password is required but not provided")
		}
		log.Println("Providing sudo password through stdin for local command")

		// Executing the command within a shell
		command := exec.Command("sudo", "-S", "bash", "-c", cmd)
		command.Stdin = strings.NewReader(sudoPassword + "\n") // Write password to stdin
		out, err := command.CombinedOutput()
		outputStr := string(out)

		// Check for sudo-related errors
		if strings.Contains(outputStr, "incorrect password") {
			return "", errors.New("sudo: incorrect password provided")
		}
		if strings.Contains(outputStr, "is not in the sudoers file") {
			return "", errors.New("sudo: user is not in the sudoers file")
		}
		if err != nil {
			log.Printf("Error running local command with sudo: %v, Output: %s\n", err, outputStr)
			return "", err
		}
		return outputStr, nil
	}

	// Executing the command within a shell for non-sudo commands as well
	command := exec.Command("bash", "-c", cmd)
	out, err := command.Output()
	if err != nil {
		log.Printf("Error running local command: %v\n", err)
		return "", err
	}
	return string(out), nil
}

func (h UnixHost) runRemoteCommand(cmd string, useSudo bool, sudoPassword string) (string, error) {
	log.Printf("Value of useSudo: %v", useSudo)
	if h.SSHClient == nil {
		return "", errors.New("SSHClient is not initialized")
	}
	config, err := h.getSSHConfig()
	if err != nil {
		return "", err
	}

	timeout := 15 * time.Minute
	client, err := h.SSHClient.Dial("tcp", h.Hostname()+":22", config, timeout)
	if err != nil || client == nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	if useSudo {
		if sudoPassword == "" {
			return "", errors.New("sudo: password is required but not provided")
		}
		cmd = "sudo -S " + cmd
		log.Println("Providing sudo password through stdin for remote command")
		session.Stdin = strings.NewReader(sudoPassword + "\n") // Write password to stdin
	}

	// Handling command timeout
	outputCh := make(chan commandResult)
	go func() {
		output, err := session.CombinedOutput(cmd)
		outputCh <- commandResult{output: output, err: err}
	}()

	select {
	case result := <-outputCh:
		outputStr := string(result.output)

		if result.err != nil {
			if exitError, ok := result.err.(*exec.ExitError); ok {
				status := exitError.Sys().(syscall.WaitStatus)
				errorMsg := fmt.Sprintf("Command '%s' over SSH failed with status %d. Output: %s", cmd, status.ExitStatus(), outputStr)

				switch status.ExitStatus() {
				case 100:
					log.Printf("Status 100 (commonly indicates a package manager error): %s", errorMsg)
				default:
					log.Printf("%s", errorMsg)
				}
				return outputStr, errors.New(errorMsg) // Ensure you're returning the detailed error
			} else {
				log.Printf("Error running command '%s' over SSH: %v", cmd, result.err)
				return outputStr, result.err
			}
		}

		// Check for sudo-related errors
		if strings.Contains(outputStr, "incorrect password") {
			return "", errors.New("sudo: incorrect password provided")
		}
		if strings.Contains(outputStr, "is not in the sudoers file") {
			return "", errors.New("sudo: user is not in the sudoers file")
		}
		return outputStr, nil

	case <-time.After(timeout):
		log.Printf("Command '%s' over SSH timed out after %v.", cmd, timeout)
		return "", errors.New("command timed out")
	}
}

func (h UnixHost) getSSHConfig() (*ssh.ClientConfig, error) {
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
			return nil, err
		}

		authMethod = ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
			return keys, nil
		})
	}

	return &ssh.ClientConfig{
		User:            h.User,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}
