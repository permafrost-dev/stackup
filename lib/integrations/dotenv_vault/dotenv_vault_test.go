package dotenvvault_test

import (
	"testing"

	dotenvvault "github.com/stackup-app/stackup/lib/integrations/dotenv_vault"
	"github.com/stackup-app/stackup/lib/settings"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocking godotenvvault.Read
type MockGodotenvvault struct {
	Env []string
	mock.Mock
}

func (m *MockGodotenvvault) Read() (map[string]string, error) {
	args := m.Called()
	return args.Get(0).(map[string]string), args.Error(1)
}

// Mocking the AppWorkflowContract for testing
type MockWorkflow struct {
	mock.Mock
	Env []string
}

// FindTaskById implements types.AppWorkflowContract.
func (*MockWorkflow) FindTaskById(id string) (any, bool) {
	panic("unimplemented")
}

// GetJsEngine implements types.AppWorkflowContract.
func (*MockWorkflow) GetJsEngine() *types.JavaScriptEngineContract {
	panic("unimplemented")
}

// GetSettings implements types.AppWorkflowContract.
func (*MockWorkflow) GetSettings() *settings.Settings {
	panic("unimplemented")
}

func (m *MockWorkflow) GetEnvSection() []string {
	return m.Env
}
func TestDotEnvVaultIntegration_Run(t *testing.T) {
	// Mocking godotenvvault
	mockGodotenvvault := new(MockGodotenvvault)

	// Test case: IsEnabled is false
	workflow := func() types.AppWorkflowContract {
		return &MockWorkflow{Env: []string{}}
	}
	integration := dotenvvault.New(func() types.AppWorkflowContract {
		return workflow()
	})

	err := integration.Run()
	assert.Nil(t, err)

	// Test case: IsEnabled is true but .env.vault file doesn't exist
	workflow = func() types.AppWorkflowContract {
		return &MockWorkflow{Env: []string{"dotenv://vault"}}
	}
	integration = dotenvvault.New(func() types.AppWorkflowContract {
		return workflow()
	})

	err = integration.Run()
	assert.Nil(t, err)

	// Test case: IsEnabled is true, .env.vault file exists, and godotenvvault.Read returns env variables
	mockGodotenvvault.On("Read").Return(map[string]string{"KEY": "VALUE"}, nil)
	err = integration.Run()
	assert.Nil(t, err)
}

func TestDotEnvVaultIntegration_Name(t *testing.T) {
	integration := dotenvvault.New(func() types.AppWorkflowContract {
		return &MockWorkflow{}
	})
	assert.Equal(t, "dotenv-vault", integration.Name())
}

func TestDotEnvVaultIntegration_IsEnabled(t *testing.T) {
	// Test case: dotenv://vault is not in the envSection
	integration := dotenvvault.New(func() types.AppWorkflowContract {
		return &MockWorkflow{
			Env: []string{},
		}
	})
	assert.False(t, integration.IsEnabled())

	// Test case: dotenv://vault is in the envSection
	integration = dotenvvault.New(func() types.AppWorkflowContract {
		return &MockWorkflow{
			Env: []string{"dotenv://vault"},
		}
	})

	assert.True(t, integration.IsEnabled())
}
