package updater

import (
	"encoding/json"

	"github.com/golang-module/carbon/v2"
	"github.com/stackup-app/stackup/lib/semver"
)

type Release struct {
	Name        string `json:"name"`
	TagName     string `json:"tag_name"`
	Prerelease  bool   `json:"prerelease"`
	PublishedAt string `json:"published_at"`
}

func NewReleaseFromJson(jsonString string) *Release {
	var release = Release{}
	json.Unmarshal([]byte(jsonString), &release)
	return &release
}

func (r *Release) TimeSinceRelease() string {
	return carbon.Parse(r.PublishedAt).DiffForHumans()
}

func (r *Release) ToJson() string {
	json, err := json.Marshal(r)
	if err != nil {
		return ""
	}

	return string(json)
}

func (r *Release) IsNewerThan(version string) bool {
	return semver.ParseSemverString(r.PublishedAt).GreaterThan(version)
}
