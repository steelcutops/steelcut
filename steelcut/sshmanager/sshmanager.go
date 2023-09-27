package sshmanager

import (
	"context"

	"github.com/steelcutops/steelcut/steelcut/commandmanager"
	"golang.org/x/crypto/ssh"
)

type SSHManager interface {
	LoadKeys(keyPassphrase string) ([]ssh.Signer, error)
	DialAndCreateSession(ctx context.Context, config *ssh.ClientConfig) (*ssh.Session, error)
	ExecuteCommand(session *ssh.Session, cmd string) (commandmanager.CommandResult, error)
}
