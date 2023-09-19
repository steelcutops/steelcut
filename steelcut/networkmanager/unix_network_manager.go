package networkmanager

import (
	"context"
	"errors"
	"fmt"
	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
	"regexp"
	"strconv"
)

type UnixNetworkManager struct {
	CommandManager cm.CommandManager
}

func (unm *UnixNetworkManager) Ping(address string) (PingResult, error) {
	// For simplicity, we'll use the 'ping' command and send a single packet
	output, err := unm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "ping",
		Args:    []string{"-c", "1", address},
	})
	if err != nil {
		return PingResult{}, err
	}

	// Parsing the 'ping' command output to extract RTT
	// Typically, the line we're interested in looks like: rtt min/avg/max/mdev = 0.029/0.029/0.029/0.000 ms
	regex := regexp.MustCompile(`rtt min/avg/max/mdev = (.+?)/(.+?)/(.+?)/(.+?) ms`)
	matches := regex.FindStringSubmatch(output.STDOUT)
	if matches == nil {
		return PingResult{}, errors.New("unable to parse ping output")
	}

	// Convert average RTT to float64
	rtt, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return PingResult{}, fmt.Errorf("unable to convert RTT to float: %v", err)
	}

	return PingResult{
		Address: address,
		RTT:     rtt,
		Success: true,
	}, nil
}
