package servicemanager

type ServiceStatus string

const (
	Active   ServiceStatus = "active"
	Inactive ServiceStatus = "inactive"
	Failed   ServiceStatus = "failed"
)

// ServiceManager represents operations that can be performed on system services.
type ServiceManager interface {
	EnableService(serviceName string) error
	DisableService(serviceName string) error
	StartService(serviceName string) error
	StopService(serviceName string) error
	RestartService(serviceName string) error
	ReloadService(serviceName string) error
	CheckServiceStatus(serviceName string) (ServiceStatus, error)
	IsServiceEnabled(serviceName string) (bool, error)
}
