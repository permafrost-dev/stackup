package updater

import (
	"fmt"
	"strings"

	"github.com/golang-module/carbon/v2"
	"github.com/stackup-app/stackup/lib/cache"
	"github.com/stackup-app/stackup/lib/gateway"
)

type Updater struct {
	cache *cache.Cache
	gw    *gateway.Gateway
}

func New(c *cache.Cache, gw *gateway.Gateway) *Updater {
	return &Updater{cache: c, gw: gw}
}

// Example: updater.New(cache, gw).IsUpdateAvailable("v0.0.1", "permafrost-dev/stackup")
func (u *Updater) IsUpdateAvailable(githubRepository string, currentVersion string) (bool, *Release) {
	cacheKey := u.cache.MakeCacheKey("latest-release", githubRepository)

	// return cached response if one exists
	if u.cache.HasUnexpired(cacheKey) {
		releaseJson, _ := u.cache.Get(cacheKey)
		release := NewReleaseFromJson(releaseJson.Value)
		return release.IsNewerThan(currentVersion), release
	}

	release, err := u.fetchLatestRepositoryRelease(githubRepository)
	if err != nil {
		return false, nil
	}

	//cache the response
	releaseJson := release.ToJson()
	if len(releaseJson) > 0 {
		expires := carbon.Now().AddHours(12)
		u.cache.Set(
			cacheKey,
			u.cache.CreateEntry(cacheKey, releaseJson, &expires, "", "", nil),
			int(carbon.Now().DiffInMinutes(expires)),
		)
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
