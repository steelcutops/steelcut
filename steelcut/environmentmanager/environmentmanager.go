package environmentmanager

type EnvironmentManager interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Unset(key string) error
	List() (map[string]string, error)
}
