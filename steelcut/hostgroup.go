package steelcut

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
	for _, host := range hg.Hosts {
		result, err := host.RunCommand(cmd)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
