package steelcut

import (
	"sync"
)

type HostGroup struct {
	Hosts []Host
}

func NewHostGroup(hosts ...Host) *HostGroup {
	return &HostGroup{Hosts: hosts}
}

func (hg *HostGroup) AddHost(host Host) {
	hg.Hosts = append(hg.Hosts, host)
}

func (hg *HostGroup) RunCommandOnAll(cmd string) ([]string, error) {
	var results []string
	var errors []error
	var wg sync.WaitGroup
	resultsChan := make(chan string, len(hg.Hosts))
	errChan := make(chan error, len(hg.Hosts))

	for _, host := range hg.Hosts {
		wg.Add(1)
		go func(h Host) {
			defer wg.Done()
			result, err := h.RunCommand(cmd)
			if err != nil {
				errChan <- err
			} else {
				resultsChan <- result
			}
		}(host)
	}

	wg.Wait()
	close(resultsChan)
	close(errChan)

	for result := range resultsChan {
		results = append(results, result)
	}

	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) != 0 {
		return nil, errors[0]
	}

	return results, nil
}
