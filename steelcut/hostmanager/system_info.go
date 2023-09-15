package hostmanager

import "time"

// SystemInfoOperations represents operations that retrieve system information.
type SystemInfoOperations interface {
	Hostname() (string, error)
	Uptime() (time.Duration, error)
	CPUCount() (int, error)
	TotalMemory() (int64, error)  // Return memory in bytes
	FreeMemory() (int64, error)   // Return free memory in bytes
}
