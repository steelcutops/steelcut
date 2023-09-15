package networkmanager

// UsageStats contains the information about data sent and received on a particular interface.
type UsageStats struct {
	InterfaceName string
	DataSent      uint64 // bytes
	DataReceived  uint64 // bytes
}

// Connection represents an active network connection.
type Connection struct {
	LocalAddress  string
	RemoteAddress string
	Protocol      string
	State         string
}

// FirewallRule represents a rule in the firewall.
type FirewallRule struct {
	RuleID      int
	Source      string
	Destination string
	Protocol    string
	Port        int
	Action      string // Allow, Deny, etc.
}

// Route represents a network route.
type Route struct {
	Destination string
	Gateway     string
	Metric      int
	Interface   string
}

// PingResult represents the result of a ping operation.
type PingResult struct {
	Address string
	RTT     float64 // Round Trip Time in milliseconds
	Success bool
}

// TraceStep represents a step in a traceroute operation.
type TraceStep struct {
	Hop  int
	Host string
	RTT  float64 // Round Trip Time in milliseconds
}

type NetworkManager interface {
	// Connectivity operations
	IsConnectedToInternet() (bool, error)
	CanReachAddress(address string) (bool, error)

	// Configuration operations
	GetCurrentIPAddress(interfaceName string) (string, error)
	SetIPAddress(interfaceName string, ipAddress string) error
	GetDNSServers() ([]string, error)
	SetDNSServers(servers []string) error
	EnableInterface(interfaceName string) error
	DisableInterface(interfaceName string) error

	// Monitoring and reporting
	GetCurrentNetworkUsage(interfaceName string) (UsageStats, error)
	ListActiveConnections() ([]Connection, error)

	// Advanced network operations
	AddFirewallRule(rule FirewallRule) error
	DeleteFirewallRule(rule FirewallRule) error
	ListRoutes() ([]Route, error)
	AddRoute(route Route) error
	DeleteRoute(route Route) error

	// Utility operations
	Ping(address string) (PingResult, error)
	TraceRoute(address string) ([]TraceStep, error)
	DNSLookup(domain string) ([]string, error)
}
