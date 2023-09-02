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

// SystemReporter defines an interface for reporting system-related information.
type SystemReporter interface {
	CPUUsage() (float64, error)
	DiskUsage() (float64, error)
	MemoryUsage() (float64, error)
	RunningProcesses() ([]string, error)
}

// Host defines an interface for performing operations on a host system.
type Host interface {
	SSHClient
	FileManager
	CommandExecutor
	PackageManager
	OSDetector
	SystemReporter
	CheckUpdates() ([]Update, error)
	Hostname() string
	IsReachable() error
	Reboot() error
	Shutdown() error
	UpgradeAllPackages() ([]Update, error)
}

type HostOptions struct {
	Client        SSHClient
	FileMngr      FileManager
	CmdExecutor   CommandExecutor
	PackageMngr   PackageManager
	OSDet         OSDetector
	SysReporter   SystemReporter
	User          string
	Password      string
	KeyPassphrase string
	OS            string
	SudoPassword  string
	HostString    string
}

type HostOption func(*HostOptions)

type CommandExecutor interface {
	RunCommand(command string, options CommandOptions) (string, error)
}

type commandResult struct {
	output []byte
	err    error
}

type OSDetector interface {
	DetermineOS(options *HostOptions) (string, error)
}

type DefaultOSDetector struct{}

func (d DefaultOSDetector) DetermineOS(options *HostOptions) (string, error) {
	output, err := options.CmdExecutor.RunCommand("uname", CommandOptions{UseSudo: false})
	if err != nil {
		return "", err
	}
	osType := strings.TrimSpace(output)

	if osType == "Linux" {
		osRelease, err := options.CmdExecutor.RunCommand("cat /etc/os-release", CommandOptions{UseSudo: false})
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

func WithKeyPassphrase(keyPassphrase string) HostOption {
	return func(options *HostOptions) {
		options.KeyPassphrase = keyPassphrase
	}
}

func WithUser(user string) HostOption {
	return func(options *HostOptions) {
		options.User = user
	}
}

func WithPassword(password string) HostOption {
	return func(options *HostOptions) {
		options.Password = password
	}
}

func WithOS(os string) HostOption {
	return func(options *HostOptions) {
		options.OS = os
	}
}

func WithSudoPassword(password string) HostOption {
	return func(options *HostOptions) {
		options.SudoPassword = password
	}
}

func NewHost(options ...HostOption) (Host, error) {
	hostOptions := &HostOptions{}

	// Apply provided options
	for _, opt := range options {
		opt(hostOptions)
	}

	// If the OS has not been specified, determine it.
	if hostOptions.OS == "" {
		osType, err := hostOptions.OSDet.DetermineOS(hostOptions)
		if err != nil {
			return nil, err
		}
		hostOptions.OS = osType
	}

	cmdOptions := CommandOptions{
		SudoPassword: hostOptions.SudoPassword,
	}

	switch {
	case isOsType(hostOptions.OS, "Linux_Ubuntu", "Linux_Debian"):
		return configureLinuxHost(hostOptions, cmdOptions, "apt"), nil
	case isOsType(hostOptions.OS, "Linux_RedHat", "Linux_CentOS"):
		return configureLinuxHost(hostOptions, cmdOptions, "yum"), nil
	case isOsType(hostOptions.OS, "Linux_Fedora"):
		return configureLinuxHost(hostOptions, cmdOptions, "dnf"), nil
	case hostOptions.OS == "Darwin":
		return configureMacHost(hostOptions, cmdOptions), nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", hostOptions.OS)
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

func configureLinuxHost(options *HostOptions, cmdOptions CommandOptions, pkgManagerType string) *LinuxHost {
	linuxHost := &LinuxHost{
		HostString: options.HostString,
	}
	if options.CmdExecutor == nil {
		options.CmdExecutor = &DefaultCommandExecutor{
			Host:    linuxHost,
			Options: cmdOptions,
		}
	}

	switch pkgManagerType {
	case "apt":
		linuxHost.PackageManager = AptPackageManager{Executor: options.CmdExecutor}
	case "yum":
		linuxHost.PackageManager = YumPackageManager{Executor: options.CmdExecutor}
	case "dnf":
		linuxHost.PackageManager = DnfPackageManager{Executor: options.CmdExecutor}
	}

	return linuxHost
}

func configureMacHost(options *HostOptions, cmdOptions CommandOptions) *MacOSHost {
	macHost := &MacOSHost{}
	if options.CmdExecutor == nil {
		options.CmdExecutor = &DefaultCommandExecutor{
			Host:    macHost,
			Options: cmdOptions,
		}
	}
	macHost.PackageManager = BrewPackageManager{Executor: options.CmdExecutor}
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
