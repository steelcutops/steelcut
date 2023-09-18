package hostmanager

import (
	"errors"
	"context"
	"strconv"
	"strings"
	"time"

	cm "github.com/steelcutops/steelcut/steelcut/commandmanager"
)

type UnixHostManager struct {
	CommandManager cm.CommandManager
}

// Info gathers comprehensive information about the host system.
func (uhm *UnixHostManager) Info() (HostInfo, error) {
	hostname, err := uhm.Hostname()
	if err != nil {
		return HostInfo{}, err
	}

	uptime, err := uhm.Uptime()
	if err != nil {
		return HostInfo{}, err
	}

	cpuCount, err := uhm.CPUCount()
	if err != nil {
		return HostInfo{}, err
	}

	kernelVersionOutput, err := uhm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "uname",
		Args:    []string{"-r"},
	})
	if err != nil {
		return HostInfo{}, err
	}

	osVersionOutput, err := uhm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "uname",
		Args:    []string{"-o"},
	})
	if err != nil {
		return HostInfo{}, err
	}

	return HostInfo{
		Hostname:       hostname,
		OSVersion:      strings.TrimSpace(osVersionOutput.STDOUT),
		KernelVersion:  strings.TrimSpace(kernelVersionOutput.STDOUT),
		Uptime:         uptime.String(),
		NumberOfCores:  cpuCount,
	}, nil
}

func (uhm *UnixHostManager) Hostname() (string, error) {
	output, err := uhm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "hostname",
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output.STDOUT), nil
}

func (uhm *UnixHostManager) Reboot() error {
	_, err := uhm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "sudo",
		Args:    []string{"reboot"},
	})
	return err
}

func (uhm *UnixHostManager) Shutdown() error {
	_, err := uhm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "sudo",
		Args:    []string{"shutdown", "-h", "now"},
	})
	return err
}

// CPUCount retrieves the number of CPU cores.
func (uhm *UnixHostManager) CPUCount() (int, error) {
	output, err := uhm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "nproc",
	})
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(strings.TrimSpace(output.STDOUT))
}

// Uptime retrieves the system's uptime duration.
func (uhm *UnixHostManager) Uptime() (time.Duration, error) {
	output, err := uhm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "uptime",
		Args:    []string{"-p"},
	})
	if err != nil {
		return 0, err
	}

	// Parse the uptime -p output. This is a naive way; you may want to enhance this parsing.
	uptimeStr := strings.TrimSpace(output.STDOUT)
	uptimeStr = strings.TrimPrefix(uptimeStr, "up ")
	uptimeArr := strings.Split(uptimeStr, ", ")
	var totalMinutes int

	for _, s := range uptimeArr {
		if strings.Contains(s, "hour") {
			hours, _ := strconv.Atoi(strings.Split(s, " ")[0])
			totalMinutes += hours * 60
		} else if strings.Contains(s, "minute") {
			minutes, _ := strconv.Atoi(strings.Split(s, " ")[0])
			totalMinutes += minutes
		}
	}

	return time.Duration(totalMinutes) * time.Minute, nil
}

// FreeMemory retrieves the amount of free memory in bytes.
func (uhm *UnixHostManager) FreeMemory() (int64, error) {
	output, err := uhm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "cat",
		Args:    []string{"/proc/meminfo"},
	})
	if err != nil {
		return 0, err
	}

	// Parse the /proc/meminfo content.
	lines := strings.Split(output.STDOUT, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemAvailable:") {
			// Extract the value and convert to bytes.
			// Assumes that the value in /proc/meminfo is in kilobytes (KB).
			parts := strings.Fields(line)
			if len(parts) < 2 {
				return 0, errors.New("unexpected format in /proc/meminfo")
			}

			kbValue, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return 0, err
			}
			
			// Convert KB to bytes
			return kbValue * 1024, nil
		}
	}

	return 0, errors.New("could not find MemAvailable in /proc/meminfo")
}

// TotalMemory retrieves the total amount of memory in bytes.
func (uhm *UnixHostManager) TotalMemory() (int64, error) {
	output, err := uhm.CommandManager.Run(context.TODO(), cm.CommandConfig{
		Command: "cat",
		Args:    []string{"/proc/meminfo"},
	})
	if err != nil {
		return 0, err
	}

	// Parse the /proc/meminfo content.
	lines := strings.Split(output.STDOUT, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			// Extract the value and convert to bytes.
			// Assumes that the value in /proc/meminfo is in kilobytes (KB).
			parts := strings.Fields(line)
			if len(parts) < 2 {
				return 0, errors.New("unexpected format in /proc/meminfo")
			}

			kbValue, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return 0, err
			}
			
			// Convert KB to bytes
			return kbValue * 1024, nil
		}
	}

	return 0, errors.New("could not find MemTotal in /proc/meminfo")
}
