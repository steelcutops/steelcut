package commandmanager

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/steelcutops/steelcut/common"
)

type MockSSHClient struct {
	dialError error
}

func (m *MockSSHClient) Dial(network, addr string, config *ssh.ClientConfig, timeout time.Duration) (*ssh.Client, error) {
	return nil, m.dialError
}

type MockSSHSession struct {
	runError error
}

func (m *MockSSHSession) Run(cmd string) error {
	return m.runError
}

func TestRunLocal(t *testing.T) {
	manager := UnixCommandManager{
		Hostname: "localhost",
		Credentials: common.Credentials{
			SudoPassword: "correct",
		},
	}

	config := CommandConfig{
		Command: "echo",
		Args:    []string{"hello"},
	}

	// Since we are not actually running the command, the actual result may not be accurate.
	// This test is mainly to check if there are any unexpected panics or errors.
	_, err := manager.RunLocal(context.Background(), config)

	if err != nil {
		t.Errorf("RunLocal failed: %v", err)
	}
}

func TestIsLocal(t *testing.T) {
	manager := UnixCommandManager{
		Hostname: "localhost",
	}

	if !manager.isLocal() {
		t.Errorf("Expected isLocal to return true for localhost")
	}

	manager.Hostname = "example.com"
	if manager.isLocal() {
		t.Errorf("Expected isLocal to return false for example.com")
	}
}

func TestRunRemoteDialError(t *testing.T) {
	manager := UnixCommandManager{
		Hostname:  "remote",
		SSHClient: &MockSSHClient{dialError: errors.New("mock dial error")},
		Credentials: common.Credentials{
			User:     "user",
			Password: "password",
		},
	}

	config := CommandConfig{
		Command: "ls",
	}

	_, err := manager.RunRemote(context.Background(), config)

	if err == nil || err.Error() != "mock dial error" {
		t.Errorf("Expected RunRemote to return mock dial error, got %v", err)
	}
}

func TestRun(t *testing.T) {
	managerLocal := UnixCommandManager{
		Hostname: "localhost",
	}
	managerRemote := UnixCommandManager{
		Hostname: "example.com",
	}

	config := CommandConfig{
		Command: "echo",
		Args:    []string{"hello"},
	}

	// Test local execution
	_, errLocal := managerLocal.Run(context.Background(), config)
	if errLocal != nil {
		t.Errorf("Run with localhost failed: %v", errLocal)
	}

	// Test remote execution (mock error for simplicity)
	_, errRemote := managerRemote.Run(context.Background(), config)
	if errRemote == nil {
		t.Errorf("Expected Run with remote host to fail due to lack of mock, but it didn't")
	}
}