package packagemanager

type PackageManager interface {
    ListPackages() ([]string, error)
    AddPackage(pkg string) error
    RemovePackage(pkg string) error
    UpgradePackage(pkg string) error
    CheckOSUpdates() ([]string, error)
    //UpgradeAll() ([]Update, error)
}
