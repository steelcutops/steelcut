package steelcut

import (
	"fmt"
	"io/ioutil"
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

	// Get the keys from the agent
	keys, err := sshAgent.List()
	if err != nil {
		return nil, fmt.Errorf("could not get keys from SSH agent: %v", err)
	}

	// Convert the keys into ssh.Signer objects
	signers := make([]ssh.Signer, len(keys))
	for i, key := range keys {
		signers[i], err = ssh.NewSignerFromKey(key)
		if err != nil {
			return nil, fmt.Errorf("could not create signer from key: %v", err)
		}
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
		keyBytes, err := ioutil.ReadFile(file)
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
