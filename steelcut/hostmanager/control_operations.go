package hostmanager

// ControlOperations represents control operations for the host.
type ControlOperations interface {
	Reboot() error
	Shutdown() error
}
