package steelcut

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCommandExecutor struct {
	mock.Mock
}

func (m *MockCommandExecutor) RunCommand(command string, asRoot bool) (string, error) {
	args := m.Called(command, asRoot)
	return args.String(0), args.Error(1)
}

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
	assert.Equal(t, []string{"package1", "package2", ""}, packages)

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
	assert.Equal(t, []string{"package1 update", "package2 update", ""}, updates)

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
