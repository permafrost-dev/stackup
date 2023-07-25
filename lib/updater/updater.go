package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Release struct {
	Name       string `json:"name"`
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
}

// Example: IsLatestApplicationReleaseNewerThanCurrent("v0.0.1", "permafrost-dev/stackup")
func IsLatestApplicationReleaseNewerThanCurrent(currentVersion string, githubRepository string) bool {
	latestRelease := getLatestApplicationRelease(githubRepository)

	latest := extractSemver(latestRelease.TagName)
	current := extractSemver(currentVersion)

	return isGreaterSemver(latest, current)
}

func getLatestApplicationRelease(repository string) *Release {
	parts := strings.Split(repository, "/")
	owner := parts[0]
	repoName := parts[1]

	release, err := getLatestReleaseForRepository(owner, repoName)
	if err != nil {
		return &Release{
			Name:       "unknown",
			TagName:    "v0.0.0",
			Prerelease: false,
		}
	}

	return release
}

func getLatestReleaseForRepository(owner, repo string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release Release

	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return nil, err
	}

	return &release, nil
}

func isGreaterSemver(version1, version2 string) bool {
	parts1 := strings.Split(version1, ".")
	parts2 := strings.Split(version2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		num1, _ := strconv.Atoi(parts1[i])
		num2, _ := strconv.Atoi(parts2[i])

		if num1 > num2 {
			return true
		} else if num1 < num2 {
			return false
		}
	}

	return len(parts1) > len(parts2)
}

func extractSemver(version string) string {
	// Match the major, minor, and patch version numbers
	re := regexp.MustCompile(`^v?(\d+\.\d+\.\d+).*`)
	matches := re.FindStringSubmatch(version)

	if len(matches) > 1 {
		return matches[1]
	}

	return version
}
