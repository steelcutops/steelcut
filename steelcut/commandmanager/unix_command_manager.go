package commandmanager

import (
	"context"
	"errors"
	"log/slog"
	"os"
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

func (u *UnixCommandManager) checkSudoErrors(result CommandResult) error {
	if strings.Contains(result.STDERR, "incorrect password") {
		return errors.New("sudo: incorrect password provided")
	}
	if strings.Contains(result.STDERR, "is not in the sudoers file") {
		return errors.New("sudo: user is not in the sudoers file")
	}
	if strings.Contains(result.STDERR, "timed out reading password") {
		return errors.New("sudo: password prompt timed out")
	}
	if strings.Contains(result.STDERR, "no tty present and no askpass program specified") {
		return errors.New("sudo: cannot prompt for password due to missing terminal or askpass program")
	}
	if strings.Contains(result.STDERR, "unknown user") {
		return errors.New("sudo: specified user is unknown")
	}
	if strings.Contains(result.STDERR, "unable to execute") {
		return errors.New("sudo: unable to execute the specified command")
	}
	if strings.Contains(result.STDERR, "Permission denied") {
		return errors.New("permission denied: consider using sudo for this command")
	}
	return nil
}

func (u *UnixCommandManager) RunLocal(ctx context.Context, config CommandConfig) (CommandResult, error) {
	start := time.Now()

	cmd := exec.CommandContext(ctx, config.Command, config.Args...)
	if config.Sudo {
		cmdArgs := append([]string{"sudo", "-S", "--", config.Command}, config.Args...)
		cmd = exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdin = strings.NewReader(u.SudoPassword + "\n")
	}

	// Set the environment variables
	if len(config.Env) > 0 {
		cmd.Env = append(os.Environ(), config.Env...)
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
	sudoErr := u.checkSudoErrors(result)
	if sudoErr != nil {
		return result, sudoErr
	}

	return result, err
}

func (c UnixCommandManager) getSSHConfig() (*ssh.ClientConfig, error) {
	var authMethods []ssh.AuthMethod

	handleKeyboardInteractive := func(user, instruction string, questions []string, echos []bool) ([]string, error) {
		for _, question := range questions {
			slog.Debug("Received keyboard-interactive challenge:", "challenge", question, "hostname", c.Hostname)
		}

		// Return an empty response to prevent hanging.
		return make([]string, len(questions)), nil
	}

	if c.Password != "" {
		slog.Debug("Using password authentication", "hostname", c.Hostname)
		authMethods = append(authMethods, ssh.Password(c.Password))
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

		authMethods = append(authMethods, ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
			return keys, nil
		}))
	}

	// Add keyboard-interactive authentication method
	authMethods = append(authMethods, ssh.KeyboardInteractive(handleKeyboardInteractive))

	return &ssh.ClientConfig{
		User:            c.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

func (u *UnixCommandManager) RunRemote(ctx context.Context, config CommandConfig) (CommandResult, error) {
	slog.Debug("Executing remote command",
		"hostname", u.Hostname,
		"command", config.Command,
		"args", strings.Join(config.Args, " "),
		"sudo", config.Sudo,
	)

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
	if err != nil || client == nil {
		return CommandResult{}, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil || session == nil {
		return CommandResult{}, err
	}
	defer session.Close()

	cmdStr := config.Command + " " + strings.Join(config.Args, " ")

	if config.Sudo {
		cmdStr = "sudo -S -- " + cmdStr
		session.Stdin = strings.NewReader(u.SudoPassword + "\n")
	}

	// Prepend environment variables
	if len(config.Env) > 0 {
		envStr := strings.Join(config.Env, " ") + " "
		cmdStr = envStr + cmdStr
	}

	start := time.Now()

	outputCh := make(chan CommandResult)
	go func() {
		var result CommandResult

		// Set up the command to execute remotely
		var stdout, stderr strings.Builder
		session.Stdout = &stdout
		session.Stderr = &stderr

		// Execute command
		err := session.Run(cmdStr)
		if err != nil {
			slog.Error("Failed to execute command over SSH", "command", cmdStr, "error", err, "stdout", stdout.String(), "stderr", stderr.String())
			result.ExitCode = getExitCode(err)
		}

		result.STDOUT = stdout.String()
		result.STDERR = stderr.String()

		// Send the result to the channel
		outputCh <- result
	}()

	select {
	case result := <-outputCh:
		result.Duration = time.Since(start)
		result.Timestamp = start
		result.Command = cmdStr

		// Check for sudo-related errors
		sudoErr := u.checkSudoErrors(result)
		if sudoErr != nil {
			return result, sudoErr
		}

		return result, nil

	case <-ctx.Done():
		slog.Error("Command over SSH timed out.", "command_string", cmdStr)
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
