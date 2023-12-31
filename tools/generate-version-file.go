package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/blang/semver"
)

const (
	VERSION     = "0.0.0"
	TARGET_FILE = "lib/version/app-version.go"
)

func getTag(match ...string) (string, *semver.PRVersion) {
	args := append([]string{
		"describe", "--tags",
	}, match...)
	tag, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", nil
	}
	tagParts := strings.Split(string(tag), "-")
	if len(tagParts) == 3 {
		if ahead, err := semver.NewPRVersion(tagParts[1]); err == nil {
			return tagParts[0], &ahead
		}
	} else if len(tagParts) == 4 {
		if ahead, err := semver.NewPRVersion(tagParts[2]); err == nil {
			return tagParts[0] + "-" + tagParts[1], &ahead
		}
	}

	return string(tag), nil
}

func getProjectVersion() string {
	if tags, err := exec.Command("git", "tag").Output(); err != nil || len(tags) == 0 {
		// no tags found -- fetch them
		exec.Command("git", "fetch", "--tags").Run()
	}
	// Find the last vX.X.X Tag and get how many builds we are ahead of it.
	versionStr, ahead := getTag("--match", "v*")
	version, err := semver.ParseTolerant(versionStr)
	if err != nil {
		// no version tag found so just return what ever we can find.
		return "0.0.0-unknown"
	}
	// Get the tag of the current revision.
	tag, _ := getTag("--exact-match")
	if tag == versionStr {
		// Seems that we are going to build a release.
		// So the version number should already be correct.
		return version.String()
	}

	// If we don't have any tag assume "dev"
	if tag == "" || strings.HasPrefix(tag, "nightly") {
		tag = "dev"
	}
	// Get the most likely next version:
	if !strings.Contains(version.String(), "rc") {
		version.Patch = version.Patch + 1
	}

	if pr, err := semver.NewPRVersion(tag); err == nil {
		// append the tag as pre-release name
		version.Pre = append(version.Pre, pr)
	}

	if ahead != nil {
		// if we know how many commits we are ahead of the last release, append that too.
		version.Pre = append(version.Pre, *ahead)
	}

	return version.String()
}

func RunGitCommand(args ...string) string {
	cmd := exec.Command("git", args...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(out.String())
}

func main() {
	result := getProjectVersion() //RunGitCommand("rev-parse", "HEAD")[0:8]

	file, err := os.Create(TARGET_FILE)
	if err != nil {
		fmt.Printf("Cannot create %s file: %v\n", TARGET_FILE, err)
		os.Exit(1)
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("package version\n\nconst APP_VERSION = \"%s\"\n", result))
}
