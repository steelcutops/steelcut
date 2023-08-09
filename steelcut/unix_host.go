package steelcut

type UnixHost struct {
	User          string
	Password      string
	KeyPassphrase string
	OS            string
	SudoPassword  string
	SSHClient     SSHClient
	HostString    string
}

func (h *UnixHost) Hostname() string {
	return h.HostString // Return the renamed field
}
