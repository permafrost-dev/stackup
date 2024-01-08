package infisical

import (
	"errors"
	"os"
	"strings"

	"github.com/stackup-app/stackup/lib/types"
)

type InfisicalIntegration struct {
	workflow func() types.AppWorkflowContract
	endpoint string
	client   *Client
}

func New(getWorkflow func() types.AppWorkflowContract) InfisicalIntegration {
	return InfisicalIntegration{workflow: getWorkflow, endpoint: ""}
}

func (in InfisicalIntegration) Name() string {
	return "infisical"
}

func (in InfisicalIntegration) IsEnabled() bool {
	for _, value := range in.workflow().GetEnvSection() {
		if strings.HasPrefix(value, "infisical://") {
			in.endpoint = strings.TrimPrefix(value, "infisical://")
			return true
		}
	}
	return false
}

// ParseInfiscalURL parses the given string in the format `infiscal://<workspace_id>:<environment_name>/<path>`
// and returns the workspace_id, environment_name, and path.
func (in InfisicalIntegration) parseInfisicalURL(infisicalUrl string) (string, string, string, error) {
	url := os.ExpandEnv(infisicalUrl)

	const prefix = "infisical://"
	if !strings.HasPrefix(url, prefix) {
		return "", "", "", errors.New("invalid prefix")
	}

	url = strings.TrimPrefix(url, prefix)
	parts := strings.SplitN(url, "/", 2)
	if len(parts) == 0 {
		return "", "", "", errors.New("invalid workspace and environment format")
	}

	if len(parts) < 2 {
		parts = append(parts, "")
	}

	workspaceAndEnv := parts[0]
	path := parts[1]

	workspaceAndEnvParts := strings.SplitN(workspaceAndEnv, ":", 2)
	if len(workspaceAndEnvParts) != 2 {
		return "", "", "", errors.New("invalid workspace and environment format")
	}

	workspaceID := workspaceAndEnvParts[0]
	environmentName := workspaceAndEnvParts[1]

	return workspaceID, environmentName, path, nil
}

func (in InfisicalIntegration) Run() error {
	if !in.IsEnabled() {
		return nil
	}

	in.client = NewClient("https://app.infisical.com")

	for _, value := range in.workflow().GetEnvSection() {
		err := in.processEnvEntry(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (in InfisicalIntegration) processEnvEntry(str string) error {
	workspaceID, environmentName, path, err := in.parseInfisicalURL(str)

	secrets, err := in.client.GetSecrets(workspaceID, environmentName, "/"+path, true)

	if err != nil {
		return err
	}

	for _, secret := range secrets {
		if secret.Name != "" {
			os.Setenv(secret.Name, secret.Value)
		}
	}

	return nil
}
