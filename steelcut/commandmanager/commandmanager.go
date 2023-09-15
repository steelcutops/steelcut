package commandmanager

// CommandResult encapsulates the results from a command execution.
type CommandResult struct {
	STDOUT   string
	STDERR   string
	ExitCode int
}

// CommandManager provides methods to execute commands, both locally and remotely.
type CommandManager interface {
	// RunLocal executes a command on the local system.
	RunLocal(command string, args ...string) (CommandResult, error)

	// RunRemote executes a command on a remote system via SSH.
	RunRemote(host, command string, args ...string) (CommandResult, error)

	// RunRemoteWithTimeout executes a command on a remote system with a timeout.
	RunRemoteWithTimeout(host, command string, timeout int, args ...string) (CommandResult, error)

	// ... Other methods as needed, like RunBatch, StreamOutput, etc.
}
