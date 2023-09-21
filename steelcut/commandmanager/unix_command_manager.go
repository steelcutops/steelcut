package commandmanager

import (
	"context"
	"errors"
	"log/slog"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/steelcutops/steelcut/common"
	"github.com/steelcutops/steelcut/steelcut"
	"golang.org/x/crypto/ssh"
)

type SSHDialer interface {
	Dial(network, addr string, config *ssh.ClientConfig, timeout time.Duration) (*ssh.Client, error)
}

type UnixCommandManager struct {
	Hostname  string
	SSHClient SSHDialer
	common.Credentials
}

func (u *UnixCommandManager) RunLocal(ctx context.Context, config CommandConfig) (CommandResult, error) {
	start := time.Now()

	cmd := exec.CommandContext(ctx, config.Command, config.Args...)
	if config.Sudo {
		cmdArgs := append([]string{"sudo", "-S", config.Command}, config.Args...)
		cmd = exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)

		cmd.Stdin = strings.NewReader(u.SudoPassword + "\n")
	}
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	duration := time.Since(start)
	result := CommandResult{
		Command:   config.Command,
		STDOUT:    stdout.String(),
		STDERR:    stderr.String(),
		ExitCode:  getExitCode(err),
		Duration:  duration,
		Timestamp: start,
	}

	// Check for sudo-related errors
	if strings.Contains(result.STDOUT, "incorrect password") {
		return result, errors.New("sudo: incorrect password provided")
	}
	if strings.Contains(result.STDOUT, "is not in the sudoers file") {
		return result, errors.New("sudo: user is not in the sudoers file")
	}

	return result, err
}

func (c UnixCommandManager) getSSHConfig() (*ssh.ClientConfig, error) {
	var authMethod ssh.AuthMethod

	if c.Password != "" {
		slog.Debug("Using password authentication", "hostname", c.Hostname)
		authMethod = ssh.Password(c.Password)
	} else {
		slog.Debug("Using public key authentication", "hostname", c.Hostname)
		var keyManager steelcut.SSHKeyManager
		if c.KeyPassphrase != "" {
			keyManager = steelcut.FileSSHKeyManager{}
		} else {
			keyManager = steelcut.AgentSSHKeyManager{}
		}

		keys, err := keyManager.ReadPrivateKeys(c.KeyPassphrase)
		if err != nil {
			return nil, err
		}

		authMethod = ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
			return keys, nil
		})
	}

	return &ssh.ClientConfig{
		User:            c.User,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

func (u *UnixCommandManager) RunRemote(ctx context.Context, config CommandConfig) (CommandResult, error) {
	slog.Debug("Executing remote command", "hostname", u.Hostname, "command", config.Command)

	if u.SSHClient == nil {
		return CommandResult{}, errors.New("SSHClient is not initialized")
	}

	sshConfig, err := u.getSSHConfig()
	if err != nil {
		return CommandResult{}, err
	}
	var dialTimeout time.Duration
	if deadline, ok := ctx.Deadline(); ok {
		dialTimeout = time.Until(deadline)
	} else {
		dialTimeout = 15 * time.Minute
	}

	client, err := u.SSHClient.Dial("tcp", u.Hostname+":22", sshConfig, dialTimeout)

	if err != nil {
		return CommandResult{}, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return CommandResult{}, err
	}
	defer session.Close()

	cmdStr := config.Command + " " + strings.Join(config.Args, " ")

	if config.Sudo {
		cmdStr = "sudo -S " + cmdStr
		session.Stdin = strings.NewReader(u.SudoPassword + "\n")
	}

	start := time.Now()

	outputCh := make(chan CommandResult)
	go func() {
		go func() {
			var result CommandResult

			// Set up the command to execute remotely
			var stdout, stderr strings.Builder
			session.Stdout = &stdout
			session.Stderr = &stderr

			// Execute command
			err := session.Run(cmdStr)
			if err != nil {
				slog.Error("Failed to execute command ver SSH", "command", cmdStr, "error", err, "stdout", stdout.String(), "stderr", stderr.String())
				result.ExitCode = getExitCode(err)
			}

			result.STDOUT = stdout.String()
			result.STDERR = stderr.String()

			// Send the result to the channel
			outputCh <- result
		}()

	}()

	select {
	case result := <-outputCh:
		result.Duration = time.Since(start)
		result.Timestamp = start
		result.Command = cmdStr

		// Check for sudo-related errors
		outputStr := result.STDOUT
		if strings.Contains(outputStr, "incorrect password") {
			return result, errors.New("sudo: incorrect password provided")
		}
		if strings.Contains(outputStr, "is not in the sudoers file") {
			return result, errors.New("sudo: user is not in the sudoers file")
		}

		return result, nil

	case <-ctx.Done():
		slog.Error("Command '%s' over SSH timed out.", cmdStr)
		return CommandResult{}, ctx.Err()
	}

}

func (u *UnixCommandManager) Run(ctx context.Context, config CommandConfig) (CommandResult, error) {
	if u.isLocal() {
		slog.Debug("Detected local so running local command", "hostname", u.Hostname, "command", config.Command, "sshclient", u.SSHClient)
		return u.RunLocal(ctx, config)
	}

	slog.Debug("Detected remote command so running remote command", "hostname", u.Hostname, "command", config.Command, "sshclient", u.SSHClient)
	return u.RunRemote(ctx, config)
}

func (u *UnixCommandManager) isLocal() bool {
	return u.Hostname == "localhost" || u.Hostname == "127.0.0.1"
}

func getExitCode(err error) int {
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			status := exitError.Sys().(syscall.WaitStatus)
			return status.ExitStatus()
		}
	}
	return 0
}
