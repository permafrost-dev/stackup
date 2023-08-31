package updater

import (
	"fmt"
	"strings"

	"github.com/stackup-app/stackup/lib/gateway"
)

type Updater struct {
	gw *gateway.Gateway
}

func New(gw *gateway.Gateway) *Updater {
	return &Updater{gw: gw}
}

// Example: updater.New(gw).IsUpdateAvailable("v0.0.1", "permafrost-dev/stackup")
func (u *Updater) IsUpdateAvailable(githubRepository string, currentVersion string) (bool, *Release) {
	release, err := u.fetchLatestRepositoryRelease(githubRepository)
	if err != nil {
		return false, nil
	}

	return release.IsNewerThan(currentVersion), release
}

func (u *Updater) fetchLatestRepositoryRelease(repository string) (*Release, error) {
	if !strings.Contains(repository, "/") {
		return nil, fmt.Errorf("invalid repository value: '%s'", repository)
	}

	body, err := u.gw.GetUrl(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repository))
	if err != nil {
		return nil, err
	}

	return NewReleaseFromJson(body), nil
}
