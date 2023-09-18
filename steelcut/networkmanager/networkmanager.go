package networkmanager

// PingResult represents the result of a ping operation.
type PingResult struct {
	Address string
	RTT     float64 // Round Trip Time in milliseconds
	Success bool
}

type NetworkManager interface {
	Ping(address string) (PingResult, error)
}
