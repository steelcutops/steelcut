package hostgroup

import (
	"context"
	"sync"

	"github.com/steelcutops/steelcut/steelcut/commandmanager"
	"github.com/steelcutops/steelcut/steelcut/host"
)

type HostGroup struct {
	sync.RWMutex
	Hosts map[string]*host.Host
}

// NewHostGroup creates a new HostGroup with the given hosts.
func NewHostGroup(hosts ...*host.Host) *HostGroup {
	hostMap := make(map[string]*host.Host)
	for _, h := range hosts {
		hostMap[h.Hostname] = h
	}
	return &HostGroup{Hosts: hostMap}
}

// AddHost adds a host to the HostGroup.
func (hg *HostGroup) AddHost(h *host.Host) {
	hg.Lock()
	defer hg.Unlock()
	hg.Hosts[h.Hostname] = h
}

// RemoveHost removes a host from the HostGroup by its hostname.
func (hg *HostGroup) RemoveHost(hostname string) {
	hg.Lock()
	defer hg.Unlock()
	delete(hg.Hosts, hostname)
}

// HasHost checks if a host with the given hostname exists in the HostGroup.
func (hg *HostGroup) HasHost(hostname string) bool {
	hg.RLock()
	defer hg.RUnlock()
	_, exists := hg.Hosts[hostname]
	return exists
}

func (hg *HostGroup) Run(ctx context.Context, cmd string, args ...string) []commandmanager.CommandResult {
	var wg sync.WaitGroup
	results := make([]commandmanager.CommandResult, len(hg.Hosts))

	config := commandmanager.CommandConfig{
		Command: cmd,
		Args:    args,
		Sudo:    false,
	}

	hg.RLock()
	i := 0
	for _, h := range hg.Hosts {
		wg.Add(1)
		go func(hostInstance *host.Host, index int) {
			defer wg.Done()
			result, err := hostInstance.CommandManager.Run(ctx, config)
			if err != nil {
				result.STDERR = err.Error()
			}
			result.Command = cmd // Store the command in the result
			results[index] = result
		}(h, i)
		i++
	}
	hg.RUnlock()

	wg.Wait()

	return results
}
