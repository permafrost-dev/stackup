package integrations

import (
	dotenvvaultintegration "github.com/stackup-app/stackup/lib/integrations/dotenv_vault"
	"github.com/stackup-app/stackup/lib/types"
)

type Integration interface {
	Name() string
	IsEnabled() bool
	Run() error
}

func List(getWorkflow func() types.AppWorkflowContract) map[string]Integration {
	dotenvvault := dotenvvaultintegration.New(getWorkflow)

	return map[string]Integration{
		dotenvvault.Name(): dotenvvault,
	}
}
