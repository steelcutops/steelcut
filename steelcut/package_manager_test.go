package steelcut

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYumPackageManager(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	logger := log.New(os.Stdout, "test: ", log.Lshortfile)

	packageManager := YumPackageManager{
		Executor: mockExecutor,
		Logger:   logger,
	}

	// Test: ListPackages
	mockExecutor.On("RunCommand", "yum list installed", false).Return("package1\npackage2\n", nil)
	packages, err := packageManager.ListPackages(nil)
	assert.Nil(t, err)
	assert.Equal(t, []string{"package1", "package2"}, packages)

	// Test: AddPackage
	mockExecutor.On("RunCommand", "yum install -y package3", true).Return("", nil)
	err = packageManager.AddPackage(nil, "package3")
	assert.Nil(t, err)

	// Test: RemovePackage
	mockExecutor.On("RunCommand", "yum remove -y package1", true).Return("", nil)
	err = packageManager.RemovePackage(nil, "package1")
	assert.Nil(t, err)

	// Test: UpgradePackage
	mockExecutor.On("RunCommand", "yum upgrade -y package2", true).Return("", nil)
	err = packageManager.UpgradePackage(nil, "package2")
	assert.Nil(t, err)

	// Test: CheckOSUpdates
	mockExecutor.On("RunCommand", "yum check-update", true).Return("package1 update\npackage2 update\n", nil)
	updates, err := packageManager.CheckOSUpdates(nil)
	assert.Nil(t, err)
	assert.Equal(t, []string{"package1 update", "package2 update"}, updates)

	// Test: UpgradeAll
	mockExecutor.On("RunCommand", "yum update -y", true).Return("package1 2.0\npackage2 3.0\n", nil)
	upgrades, err := packageManager.UpgradeAll(nil)
	assert.Nil(t, err)
	assert.Equal(t, []Update{
		{
			PackageName: "package1",
			Version:     "2.0",
		},
		{
			PackageName: "package2",
			Version:     "3.0",
		},
	}, upgrades)

}
func TestAptPackageManager(t *testing.T) {
	mockExecutor := &MockCommandExecutor{}
	logger := log.New(os.Stdout, "test: ", log.Lshortfile)

	packageManager := AptPackageManager{
		Executor: mockExecutor,
		Logger:   logger,
	}

	// Test: ListPackages
	mockExecutor.On("RunCommand", "apt list --installed", false).Return("package1/now 1.0 amd64\npackage2/now 2.0 amd64\n", nil)
	packages, err := packageManager.ListPackages(nil)
	assert.Nil(t, err)
	assert.Equal(t, []string{"package1/now 1.0 amd64", "package2/now 2.0 amd64"}, packages)

	// Test: AddPackage
	mockExecutor.On("RunCommand", "apt install -y package3", true).Return("", nil)
	err = packageManager.AddPackage(nil, "package3")
	assert.Nil(t, err)

	// Test: RemovePackage
	mockExecutor.On("RunCommand", "apt remove -y package1", true).Return("", nil)
	err = packageManager.RemovePackage(nil, "package1")
	assert.Nil(t, err)

	// Test: UpgradePackage
	mockExecutor.On("RunCommand", "apt upgrade -y package2", true).Return("", nil)
	err = packageManager.UpgradePackage(nil, "package2")
	assert.Nil(t, err)

	// Test: CheckOSUpdates
	mockExecutor.On("RunCommand", "apt update", true).Return("", nil)
	mockExecutor.On("RunCommand", "apt list --upgradable", false).Return("package1/xenial 2.0 amd64 [upgradable from: 1.0]\npackage2/xenial 3.0 amd64 [upgradable from: 2.0]", nil)
	updates, err := packageManager.CheckOSUpdates(nil)
	assert.Nil(t, err)
	assert.Equal(t, []string{"package1/xenial 2.0 amd64 [upgradable from: 1.0]", "package2/xenial 3.0 amd64 [upgradable from: 2.0]"}, updates)

}

func TestBrewPackageManager(t *testing.T) {
	mockExecutor := new(MockCommandExecutor)
	logger := log.New(os.Stdout, "test: ", log.Lshortfile)

	packageManager := BrewPackageManager{
		Executor: mockExecutor,
		Logger:   logger,
	}

	// Test: ListPackages
	mockExecutor.On("RunCommand", "brew list --version", false).Return("package1 1.0\npackage2 2.0\n", nil)
	packages, err := packageManager.ListPackages(nil)
	assert.Nil(t, err)
	assert.Equal(t, []string{"package1 1.0", "package2 2.0"}, packages)

	// Test: AddPackage
	mockExecutor.On("RunCommand", "brew install package3", false).Return("", nil)
	err = packageManager.AddPackage(nil, "package3")
	assert.Nil(t, err)

	// Test: RemovePackage
	mockExecutor.On("RunCommand", "brew uninstall package1", false).Return("", nil)
	err = packageManager.RemovePackage(nil, "package1")
	assert.Nil(t, err)

	// Test: UpgradePackage
	mockExecutor.On("RunCommand", "brew upgrade package2", false).Return("", nil)
	err = packageManager.UpgradePackage(nil, "package2")
	assert.Nil(t, err)

	// Test: CheckOSUpdates
	mockExecutor.On("RunCommand", "brew outdated", false).Return("package1 2.0\npackage2 3.0\n", nil)
	updates, err := packageManager.CheckOSUpdates(nil)
	assert.Nil(t, err)
	assert.Equal(t, []string{"package1 2.0", "package2 3.0"}, updates)

	// Test: UpgradeAll
	mockExecutor.On("RunCommand", "brew upgrade", false).Return("package1 2.0\npackage2 3.0\n", nil)
	upgrades, err := packageManager.UpgradeAll(nil)
	assert.Nil(t, err)
	assert.Equal(t, []Update{
		{
			PackageName: "package1",
			Version:     "2.0",
		},
		{
			PackageName: "package2",
			Version:     "3.0",
		},
	}, upgrades)
}
