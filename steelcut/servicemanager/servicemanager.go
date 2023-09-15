package servicemanager

// ServiceManager represents operations that can be performed on system services.
type ServiceManager interface {
	EnableService(serviceName string) error
	StartService(serviceName string) error
	StopService(serviceName string) error
	RestartService(serviceName string) error
	CheckServiceStatus(serviceName string) (string, error)
}
