package main

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestReadHostsFromFile(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := ioutil.TempFile("", "test.ini")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write a sample INI content to the file
	content := `[group1]
host1=127.0.0.1
host2=127.0.0.2

[group2]
host3=127.0.0.3`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Expected result
	expected := map[string][]string{
		"group1": {"127.0.0.1", "127.0.0.2"},
		"group2": {"127.0.0.3"},
	}

	// Read the hosts from the file
	hosts, err := readHostsFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Error reading hosts from file: %v", err)
	}

	if !reflect.DeepEqual(hosts, expected) {
		t.Errorf("Expected %v, got %v", expected, hosts)
	}
}
