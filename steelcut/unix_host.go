package steelcut

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
)

// UnixHost represents a Unix-based host system with details like username, password, and connection information.
type UnixHost struct {
	User          string
	Password      string
	KeyPassphrase string
	OS            string
	SudoPassword  string
	SSHClient     SSHClient
	HostString    string
	FileManager   FileManager
	Executor      CommandExecutor
	PkgManager    PackageManager
	Detector      OSDetector
}

// Hostname returns the host string (e.g., IP or domain name) for the Unix host.
func (h *UnixHost) Hostname() string {
	return h.HostString // Return the renamed field
}

// CreateDirectory creates a directory at the given path.
func (h UnixHost) CreateDirectory(path string) error {
	_, err := h.runCommandInternal(fmt.Sprintf("mkdir -p %s", path), false, "")
	return err
}

// DeleteDirectory deletes the directory at the given path, including all contents.
func (h UnixHost) DeleteDirectory(path string) error {
	_, err := h.runCommandInternal(fmt.Sprintf("rm -rf %s", path), false, "")
	return err
}

// ListDirectory lists the contents of the directory at the given path.
func (h UnixHost) ListDirectory(path string) ([]string, error) {
	output, err := h.runCommandInternal(fmt.Sprintf("ls %s", path), false, "")
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(output), "\n"), nil
}

// SetPermissions sets the file permissions for the given path, using the provided FileMode.
func (h UnixHost) SetPermissions(path string, mode os.FileMode) error {
	_, err := h.runCommandInternal(fmt.Sprintf("chmod %o %s", mode, path), false, "")
	return err
}

// GetPermissions retrieves the file permissions for the given path, returning them as a FileMode value.
func (h UnixHost) GetPermissions(path string) (os.FileMode, error) {
	output, err := h.runCommandInternal(fmt.Sprintf("stat -c %%a %s", path), false, "")
	if err != nil {
		return 0, err
	}
	mode, err := strconv.ParseUint(strings.TrimSpace(output), 8, 32)
	if err != nil {
		return 0, err
	}
	return os.FileMode(mode), nil
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
