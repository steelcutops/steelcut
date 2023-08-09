package steelcut

import (
	"sync"
)

type HostGroup struct {
	sync.RWMutex // Embedding a RWMutex to provide locking
	Hosts        map[string]Host
}

func NewHostGroup(hosts ...Host) *HostGroup {
	hostMap := make(map[string]Host)
	for _, host := range hosts {
		hostMap[host.Hostname()] = host
	}
	return &HostGroup{Hosts: hostMap}
}

func (hg *HostGroup) AddHost(host Host) {
	hg.Lock()
	defer hg.Unlock()
	hg.Hosts[host.Hostname()] = host
}

type CommandResult struct {
	Result string
	Error  error
	Host   string
}

func (hg *HostGroup) RunCommandOnAll(cmd string) []CommandResult {
	var wg sync.WaitGroup
	results := make([]CommandResult, len(hg.Hosts))

	hg.RLock()
	i := 0
	for _, host := range hg.Hosts {
		wg.Add(1)
		go func(h Host, index int) {
			defer wg.Done()
			result, err := h.RunCommand(cmd)
			results[index] = CommandResult{Result: result, Error: err, Host: h.Hostname()}
		}(host, i)
		i++
	}
	hg.RUnlock()

	wg.Wait()

	return results
}
