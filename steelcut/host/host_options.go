package host

type HostOption func(*Host)

// WithUser returns a HostOption that sets the user for a Host.
func WithUser(user string) HostOption {
	return func(host *Host) {
		host.User = user
	}
}

// WithPassword returns a HostOption that sets the password for a Host.
func WithPassword(password string) HostOption {
	return func(host *Host) {
		host.Password = password
	}
}

// WithKeyPassphrase returns a HostOption that sets the key passphrase for a Host.
func WithKeyPassphrase(keyPassphrase string) HostOption {
	return func(host *Host) {
		host.KeyPassphrase = keyPassphrase
	}
}

// WithOS returns a HostOption that sets the OS for a Host.
func WithOS(os OSType) HostOption {
	return func(host *Host) {
		host.OSType = os
	}
}

// WithSudoPassword returns a HostOption that sets the sudo password for a Host.
func WithSudoPassword(password string) HostOption {
	return func(host *Host) {
		host.SudoPassword = password
	}
}

