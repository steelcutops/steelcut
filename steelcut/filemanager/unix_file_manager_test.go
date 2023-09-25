package filemanager

import (
	"context"
	"errors"
	"testing"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type MockCommandManager struct {
	Result cm.CommandResult
	Err    error
}

func (m *MockCommandManager) RunLocal(ctx context.Context, config cm.CommandConfig) (cm.CommandResult, error) {
	return m.Result, m.Err
}

func (m *MockCommandManager) RunRemote(ctx context.Context, config cm.CommandConfig) (cm.CommandResult, error) {
	return m.Result, m.Err
}

func (m *MockCommandManager) Run(ctx context.Context, config cm.CommandConfig) (cm.CommandResult, error) {
	return m.Result, m.Err
}

func TestCreateDirectory(t *testing.T) {
	mockCmd := &MockCommandManager{
		Result: cm.CommandResult{},
		Err:    nil,
	}
	manager := UnixFileManager{
		CommandManager: mockCmd,
	}

	err := manager.CreateDirectory("/path/to/directory")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestDeleteDirectory(t *testing.T) {
	mockCmd := &MockCommandManager{
		Result: cm.CommandResult{},
		Err:    nil,
	}
	manager := UnixFileManager{
		CommandManager: mockCmd,
	}

	err := manager.DeleteDirectory("/path/to/directory")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestListDirectoryError(t *testing.T) {
	mockCmd := &MockCommandManager{
		Result: cm.CommandResult{},
		Err:    errors.New("mock error"),
	}
	manager := UnixFileManager{
		CommandManager: mockCmd,
	}

	_, err := manager.ListDirectory("/path/to/directory")
	if err == nil || err.Error() != "mock error" {
		t.Errorf("Expected mock error, got: %v", err)
	}
}
