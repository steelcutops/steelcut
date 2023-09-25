package hostmanager

import (
	"context"
	"testing"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type MockCommandManager struct {
	Outputs map[string]string
	Err     error
}

func (m *MockCommandManager) getMockOutput(command string) cm.CommandResult {
	if output, exists := m.Outputs[command]; exists {
		return cm.CommandResult{STDOUT: output}
	}
	return cm.CommandResult{}
}

func (m *MockCommandManager) RunLocal(ctx context.Context, config cm.CommandConfig) (cm.CommandResult, error) {
	return m.getMockOutput(config.Command), m.Err
}

func (m *MockCommandManager) RunRemote(ctx context.Context, config cm.CommandConfig) (cm.CommandResult, error) {
	return m.getMockOutput(config.Command), m.Err
}

func (m *MockCommandManager) Run(ctx context.Context, config cm.CommandConfig) (cm.CommandResult, error) {
	return m.getMockOutput(config.Command), m.Err
}

func TestInfo(t *testing.T) {
	mockCmd := &MockCommandManager{
		Outputs: map[string]string{
			"hostname": "test-hostname\n",
			"nproc":    "4\n",
			"uname":    "test-version",
		},
		Err: nil,
	}
	hostManager := UnixHostManager{
		CommandManager: mockCmd,
	}

	_, err := hostManager.Info()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestHostname(t *testing.T) {
	mockCmd := &MockCommandManager{
		Outputs: map[string]string{
			"hostname": "test-hostname\n",
		},
		Err: nil,
	}
	hostManager := UnixHostManager{
		CommandManager: mockCmd,
	}

	hostname, err := hostManager.Hostname()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if hostname != "test-hostname" {
		t.Errorf("Expected hostname 'test-hostname', got: %v", hostname)
	}
}

func TestCPUCount(t *testing.T) {
	mockCmd := &MockCommandManager{
		Outputs: map[string]string{
			"nproc": "4\n",
		},
		Err: nil,
	}
	hostManager := UnixHostManager{
		CommandManager: mockCmd,
	}

	cpuCount, err := hostManager.CPUCount()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if cpuCount != 4 {
		t.Errorf("Expected 4 CPU cores, got: %v", cpuCount)
	}
}
