package steelcut

import (
	"testing"
)

func TestNewHost(t *testing.T) {
	host, err := NewHost("localhost", WithOS("Linux"))
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
	host, err := NewHost("localhost", WithOS("Darwin"))
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
	_, err := NewHost("localhost", WithOS("UnsupportedOS"))
	if err == nil {
		t.Fatalf("Expected an error for unsupported OS, got nil")
	}

	expectedErr := "unsupported operating system: UnsupportedOS"
	if err.Error() != expectedErr {
		t.Errorf("Expected error message '%s', got: %v", expectedErr, err)
	}
}
