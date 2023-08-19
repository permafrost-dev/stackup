package updater

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-module/carbon/v2"
	"github.com/stackup-app/stackup/lib/cache"
	"github.com/stackup-app/stackup/lib/gateway"
)

type Updater struct {
	gw *gateway.Gateway
}

func New(gw *gateway.Gateway) *Updater {
	return &Updater{
		gw: gw,
	}
}

// Example: updater.New(gw).IsLatestApplicationReleaseNewerThanCurrent("v0.0.1", "permafrost-dev/stackup")
func (u *Updater) IsLatestApplicationReleaseNewerThanCurrent(c *cache.Cache, currentVersion string, githubRepository string) (bool, *Release) {
	var release *Release
	cacheKey := c.MakeCacheKey("latest-release", githubRepository)

	// return cached response
	if c.HasUnexpired(cacheKey) {
		releaseJson, _ := c.Get(cacheKey)
		release = NewReleaseFromJson(releaseJson.Value)
		return release.IsNewerThan(currentVersion), release
	}

	release = u.getLatestApplicationRelease(githubRepository)

	//cache the response
	releaseJson := release.ToJson()
	if len(releaseJson) > 0 {
		expires := carbon.Now().AddHours(3)
		c.Set(
			cacheKey,
			c.CreateEntry(cacheKey, releaseJson, &expires, "", "", nil),
			int(expires.DiffInMinutes(carbon.Now())),
		)
	}

	return release.IsNewerThan(currentVersion), release
}

func (u *Updater) getLatestApplicationRelease(repository string) *Release {
	parts := strings.Split(repository, "/")

	release, err := u.getLatestReleaseForRepository(parts[0], parts[1])
	if err != nil {
		return &Release{
			Name:       "unknown",
			TagName:    "v0.0.0",
			Prerelease: false,
		}
	}

	return release
}

func (u *Updater) getLatestReleaseForRepository(owner, repo string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	var release Release

	body, err := u.gw.GetUrl(url)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(body), &release); err != nil {
		return nil, err
	}

	return &release, nil
}
