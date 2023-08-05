package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-module/carbon/v2"
	"github.com/stackup-app/stackup/lib/cache"
)

type Release struct {
	Name        string `json:"name"`
	TagName     string `json:"tag_name"`
	Prerelease  bool   `json:"prerelease"`
	PublishedAt string `json:"published_at"`

	DaysSinceRelease  int
	HoursSinceRelease int
	TimeSinceRelease  string
}

func GetUpdateCheckUrlFormat() string {
	return "https://api.github.com/repos/%s/%s/releases/latest"
}

// Example: IsLatestApplicationReleaseNewerThanCurrent("v0.0.1", "permafrost-dev/stackup")
func IsLatestApplicationReleaseNewerThanCurrent(c *cache.Cache, currentVersion string, githubRepository string) (bool, *Release) {
	if c.Has("latest-release:"+githubRepository) && !c.IsExpired("latest-release:"+githubRepository) {
		var release = Release{}
		releaseJson, found := c.Get("latest-release:" + githubRepository)
		if found {
			json.Unmarshal([]byte(releaseJson.Value), &release)
			latest := extractSemver(release.TagName)
			current := extractSemver(currentVersion)

			release.TimeSinceRelease = carbon.Parse(release.PublishedAt).DiffForHumans()
			return isGreaterSemver(latest, current), &release
		}
	}

	latestRelease := getLatestApplicationRelease(githubRepository)

	latestRelease.TimeSinceRelease = carbon.Parse(latestRelease.PublishedAt).DiffForHumans()

	//cache the response
	latestReleaseJson, _ := json.Marshal(latestRelease)
	expires := carbon.Now().AddMinutes(60)
	updatedAt := carbon.Now()
	entry := cache.CreateCacheEntry("latest-release"+githubRepository, string(latestReleaseJson), &expires, "", "", &updatedAt)
	c.Set("latest-release:"+githubRepository, entry, 60)

	latest := extractSemver(latestRelease.TagName)
	current := extractSemver(currentVersion)

	return isGreaterSemver(latest, current), latestRelease
}

func getLatestApplicationRelease(repository string) *Release {
	parts := strings.Split(repository, "/")
	owner := parts[0]
	repoName := parts[1]

	release, err := getLatestReleaseForRepository(owner, repoName)
	if err != nil {
		return &Release{
			Name:             "unknown",
			TagName:          "v0.0.0",
			Prerelease:       false,
			DaysSinceRelease: 0,
		}
	}

	release.DaysSinceRelease, _ = daysSinceDate(release.PublishedAt)
	release.HoursSinceRelease, _ = hoursSinceDate(release.PublishedAt)

	if release.DaysSinceRelease == 0 {
		release.TimeSinceRelease = fmt.Sprintf("%d hours ago", release.HoursSinceRelease)
	} else {
		release.TimeSinceRelease = fmt.Sprintf("%d days ago", release.DaysSinceRelease)
	}

	return release
}

func getLatestReleaseForRepository(owner, repo string) (*Release, error) {
	url := fmt.Sprintf(GetUpdateCheckUrlFormat(), owner, repo)

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

func daysSinceDate(dateString string) (int, error) {
	layout := "2006-01-02T15:04:05Z"
	parsedDate, err := time.Parse(layout, dateString)
	if err != nil {
		return 0, err
	}

	currentTime := time.Now()
	difference := currentTime.Sub(parsedDate)
	days := int(difference.Hours() / 24)
	return days, nil
}

func hoursSinceDate(dateString string) (int, error) {
	layout := "2006-01-02T15:04:05Z"
	parsedDate, err := time.Parse(layout, dateString)
	if err != nil {
		return 0, err
	}

	currentTime := time.Now()
	difference := currentTime.Sub(parsedDate)
	hours := int(difference.Hours())
	return hours, nil
}
