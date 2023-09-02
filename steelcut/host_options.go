package steelcut

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
func WithOS(os OSType) HostOption {
	return func(host *UnixHost) {
		host.OSType = os
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
