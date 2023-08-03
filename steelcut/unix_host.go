package steelcut

type UnixHost struct {
	Hostname      string
	User          string
	Password      string
	KeyPassphrase string
	OS            string
	SSHClient     SSHClient
}
