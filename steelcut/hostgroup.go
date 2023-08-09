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

func (hg *HostGroup) RunCommandOnAll(cmd string) ([]string, []error) {
	var wg sync.WaitGroup
	results := make(chan string, len(hg.Hosts))
	errors := make(chan error, len(hg.Hosts))

	hg.RLock() // Lock for reading
	for _, host := range hg.Hosts {
		wg.Add(1)
		go func(h Host) {
			defer wg.Done()
			result, err := h.RunCommand(cmd)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}(host)
	}
	hg.RUnlock()

	wg.Wait()

	// Close the channels after all goroutines are done
	close(results)
	close(errors)

	// Convert channels to slices
	var res []string
	for result := range results {
		res = append(res, result)
	}

	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	return res, errs
}
