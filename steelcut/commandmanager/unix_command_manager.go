package commandmanager

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/steelcutops/steelcut/logger"
	"github.com/steelcutops/steelcut/steelcut"
	"golang.org/x/crypto/ssh"
)

var log = logger.New()

type SSHDialer interface {
	Dial(network, addr string, config *ssh.ClientConfig, timeout time.Duration) (*ssh.Client, error)
}

type UnixCommandManager struct {
	Hostname      string
	SSHClient     SSHDialer
	Password      string
	KeyPassphrase string
	User          string
}

func (u *UnixCommandManager) RunLocal(ctx context.Context, config CommandConfig) (CommandResult, error) {
	start := time.Now()

	cmd := exec.CommandContext(ctx, config.Command, config.Args...)
	if config.Sudo {
		cmd = exec.CommandContext(ctx, "sudo", append([]string{config.Command}, config.Args...)...)
	}
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	duration := time.Since(start)
	return CommandResult{
		Command:   config.Command,
		STDOUT:    stdout.String(),
		STDERR:    stderr.String(),
		ExitCode:  getExitCode(err),
		Duration:  duration,
		Timestamp: start,
	}, err
}

func (c UnixCommandManager) getSSHConfig() (*ssh.ClientConfig, error) {
	var authMethod ssh.AuthMethod

	if c.Password != "" {
		log.Debug("Using password authentication")
		authMethod = ssh.Password(c.Password)
	} else {
		log.Debug("Using public key authentication")
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

func (u *UnixCommandManager) RunRemote(ctx context.Context, host string, config CommandConfig) (CommandResult, error) {
	log.Debug("Executing remote command on host: %s", host)

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

	client, err := u.SSHClient.Dial("tcp", host+":22", sshConfig, dialTimeout)

	if err != nil {
		return CommandResult{}, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return CommandResult{}, err
	}
	defer session.Close()

	cmdStr := config.Command
	if config.Sudo {
		cmdStr = "sudo " + cmdStr
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
				log.Error("Failed to run command '%s' over SSH: %v", cmdStr, err)
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
		return result, nil

	case <-ctx.Done():
		log.Error("Command '%s' over SSH timed out.", cmdStr)
		return CommandResult{}, ctx.Err()
	}
}

func (u *UnixCommandManager) Run(ctx context.Context, host string, config CommandConfig) (CommandResult, error) {
	if u.isLocal() {
		return u.RunLocal(ctx, config)
	}
	return u.RunRemote(ctx, host, config)
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
