package hostmanager

// HostManager encompasses operations related to host management.
type HostManager interface {
	SystemInfoOperations
	ControlOperations
}
