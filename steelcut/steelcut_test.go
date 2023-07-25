package steelcut

import (
	"testing"
)

func TestNewHost(t *testing.T) {
	host, err := NewHost("localhost", WithOS("Linux"))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	linuxHost, ok := host.(LinuxHost)
	if !ok {
		t.Fatalf("Expected a LinuxHost, got: %T", host)
	}

	if linuxHost.Hostname != "localhost" {
		t.Errorf("Expected hostname to be 'localhost', got: %s", linuxHost.Hostname)
	}
}
