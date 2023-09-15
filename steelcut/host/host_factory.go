package host

func NewHost() (HostInterface, error) {
	osType, err := DetermineOS()
	if err != nil {
		return nil, err
	}

	var ch ConcreteHost

	switch osType {
	case LinuxUbuntu, LinuxDebian, LinuxFedora, LinuxRedHat, LinuxCentOS, LinuxArch, LinuxOpenSUSE:
		ch = configureLinuxHost()
	case Darwin:
		ch = configureMacHost()
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", osType)
	}

	return &ch, nil
}

func configureLinuxHost() ConcreteHost {
	return ConcreteHost{
		CommandManager: &LinuxCommandManager{},
		FileManager:    &LinuxFileManager{},
		HostManager:    &LinuxHostManager{},
		NetworkManager: &LinuxNetworkManager{},
		ServiceManager: &LinuxServiceManager{},
		PackageManager: &LinuxPackageManager{},
	}
}

func configureMacHost() ConcreteHost {
	return ConcreteHost{
		CommandManager: &DarwinCommandManager{},
		FileManager:    &DarwinFileManager{},
		HostManager:    &DarwinHostManager{},
		NetworkManager: &DarwinNetworkManager{},
		ServiceManager: &DarwinServiceManager{},
		PackageManager: &DarwinPackageManager{},
	}
}
