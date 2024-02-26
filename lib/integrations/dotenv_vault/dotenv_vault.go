package dotenvvault

import (
	"os"

	"github.com/dotenv-org/godotenvvault"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/utils"
)

type DotEnvVaultIntegration struct {
	workflow func() types.AppWorkflowContract
}

func New(getWorkflow func() types.AppWorkflowContract) DotEnvVaultIntegration {
	return DotEnvVaultIntegration{workflow: getWorkflow}
}

func (in DotEnvVaultIntegration) Name() string {
	return "dotenv-vault"
}

func (in DotEnvVaultIntegration) IsEnabled() bool {
	return utils.ArrayContains(in.workflow().GetEnvSection(), "dotenv://vault")
}

func (in DotEnvVaultIntegration) Run() error {
	if !in.IsEnabled() {
		return nil
	}

	if !utils.IsFile(utils.WorkingDir(".env.vault")) {
		return nil
	}

	vars, err := godotenvvault.Read()
	if err != nil {
		return err
	}

	for k, v := range vars {
		os.Setenv(k, v)
	}

	return nil
}
