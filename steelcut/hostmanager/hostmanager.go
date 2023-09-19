package hostmanager

import "time"

type HostInfo struct {
	Hostname      string
	OSVersion     string
	KernelVersion string
	Uptime        string
	NumberOfCores int
}

// HostManager encompasses operations related to host management.
type HostManager interface {
	Info() (HostInfo, error)
	Hostname() (string, error)
	Uptime() (time.Duration, error)
	CPUCount() (int, error)
	TotalMemory() (int64, error) // Return memory in bytes
	FreeMemory() (int64, error)  // Return free memory in bytes
	Reboot() error
	Shutdown() error
	CPUUsage() (float64, error)   // Return CPU usage as a percentage
	Processes() ([]string, error) // Return a list of running processes
}
