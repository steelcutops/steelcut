package commandmanager

import (
	"context"
	"time"
)

// CommandResult encapsulates the results from a command execution.
type CommandResult struct {
	STDOUT    string
	STDERR    string
	ExitCode  int
	Duration  time.Duration
	Command   string
	Timestamp time.Time
}

// CommandConfig holds configurations for command execution.
type CommandConfig struct {
	Command string
	Args    []string
	Sudo    bool
	Env     []string
}

// CommandManager provides methods to execute commands, both locally and remotely.
type CommandManager interface {
	// RunLocal executes a command on the local system.
	RunLocal(ctx context.Context, config CommandConfig) (CommandResult, error)

	// RunRemote executes a command on a remote system via SSH.
	RunRemote(ctx context.Context, config CommandConfig) (CommandResult, error)

	// Run executes a command on the local system if the host is localhost, otherwise it executes the command on the remote system.
	Run(ctx context.Context, config CommandConfig) (CommandResult, error)
}
