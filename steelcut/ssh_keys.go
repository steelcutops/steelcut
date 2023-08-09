package steelcut

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SSHKeyManager interface {
	ReadPrivateKeys(keyPassphrase string) ([]ssh.Signer, error)
}

type FileSSHKeyManager struct{}

type AgentSSHKeyManager struct{}

func (km AgentSSHKeyManager) ReadPrivateKeys(_ string) ([]ssh.Signer, error) {
	// Get the SSH_AUTH_SOCK environment variable
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK not set")
	}

	// Dial the SSH agent
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("could not connect to SSH agent: %v", err)
	}
	defer conn.Close()

	// Create a new SSH agent client
	sshAgent := agent.NewClient(conn)

	// Get the signers from the agent
	signers, err := sshAgent.Signers()
	if err != nil {
		return nil, fmt.Errorf("could not get signers from SSH agent: %v", err)
	}

	return signers, nil
}

func (km FileSSHKeyManager) ReadPrivateKeys(keyPassphrase string) ([]ssh.Signer, error) {
	// Find possible key files
	files, err := filepath.Glob(os.Getenv("HOME") + "/.ssh/id_*")
	if err != nil {
		return nil, err
	}

	signers := []ssh.Signer{}

	// Try each key file
	for _, file := range files {
		// Skip public keys
		if strings.HasSuffix(file, ".pub") {
			continue
		}

		// Read private key file
		keyBytes, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		var signer ssh.Signer

		// Parse private key
		if keyPassphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(keyPassphrase))
			if err != nil {
				// Failed to parse with passphrase, try next key
				continue
			}
		} else {
			signer, err = ssh.ParsePrivateKey(keyBytes)
			if err != nil {
				// Failed to parse without passphrase, try next key
				continue
			}
		}

		signers = append(signers, signer)
	}

	if len(signers) == 0 {
		// We didn't manage to parse any key files
		return nil, err
	}

	return signers, nil
}
