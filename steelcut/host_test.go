package steelcut

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/ssh"
)

type MockCommandExecutor struct {
	mock.Mock
	commandOutput string
	err           error
}

func (m *MockCommandExecutor) RunCommand(command string, options CommandOptions) (string, error) {
	called := m.MethodCalled("RunCommand", command, options.UseSudo)
	if len(called) > 1 {
		return strings.TrimSpace(called.String(0)), called.Error(1) // Trim output
	}
	return strings.TrimSpace(m.commandOutput), m.err // Trim output
}

func (m *MockCommandExecutor) SetMockResponse(output string, err error) {
	m.commandOutput = output
	m.err = err
}

type MockOSDetector struct {
	OS  string
	Err error
}

func (m MockOSDetector) DetermineOS(host *UnixHost) (string, error) {
	return m.OS, m.Err
}

func TestNewHost_InvalidHostname(t *testing.T) {
	_, err := NewHost("!@#")
	if err == nil {
		t.Fatalf("Expected an error for invalid hostname, got nil")
	}
}

type MockSSHClient struct {
	dialErr error
}

func (m *MockSSHClient) Dial(network, addr string, config *ssh.ClientConfig, timeout time.Duration) (*ssh.Client, error) {
	return nil, m.dialErr
}

func TestRunCommand(t *testing.T) {
	t.Run("Successful run local command", func(t *testing.T) {
		host, _ := NewHost("localhost", WithCommandExecutor(&MockCommandExecutor{
			commandOutput: "Success",
			err:           nil,
		}))

		commandOptions := CommandOptions{UseSudo: false}
		output, err := host.RunCommand("echo Success", commandOptions)
		assert.NoError(t, err)
		assert.Equal(t, "Success", strings.TrimSpace(output))

	})
}

func TestNewHost(t *testing.T) {
	host, err := NewHost("localhost", WithOSDetector(MockOSDetector{OS: "Linux_RedHat"}))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	linuxHost, ok := host.(*LinuxHost)
	if !ok {
		t.Fatalf("Expected a *LinuxHost, got: %T", host)
	}

	if linuxHost.Hostname() != "localhost" {
		t.Errorf("Expected hostname to be 'localhost', got: %s", linuxHost.Hostname())
	}
}

func TestNewHost_MacOS(t *testing.T) {
	host, err := NewHost("localhost", WithOSDetector(MockOSDetector{OS: "Darwin"}))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	macOSHost, ok := host.(*MacOSHost)
	if !ok {
		t.Fatalf("Expected a *MacOSHost, got: %T", host)
	}

	if macOSHost.Hostname() != "localhost" {
		t.Errorf("Expected hostname to be 'localhost', got: %s", macOSHost.Hostname())
	}
}

func TestNewHost_InvalidOS(t *testing.T) {
	_, err := NewHost("localhost", WithOSDetector(MockOSDetector{OS: "UnsupportedOS"}))
	if err == nil {
		t.Fatalf("Expected an error for unsupported OS, got nil")
	}

	expectedErr := "unsupported operating system: UnsupportedOS"
	if err.Error() != expectedErr {
		t.Errorf("Expected error message '%s', got: %v", expectedErr, err)
	}
}

func TestNewHost_NoOptions(t *testing.T) {
	_, err := NewHost("", WithOSDetector(MockOSDetector{OS: ""}))
	if err == nil {
		t.Fatalf("Expected an error for no options, got nil")
	}
}

func TestNewHost_MultipleOptions(t *testing.T) {
	host, err := NewHost("localhost", WithOSDetector(MockOSDetector{OS: "Linux_Debian"}))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	linuxHost, ok := host.(*LinuxHost)
	if !ok {
		t.Fatalf("Expected a *LinuxHost, got: %T", host)
	}

	if linuxHost.Hostname() != "localhost" {
		t.Errorf("Expected hostname to be 'localhost', got: %s", linuxHost.Hostname())
	}
}
