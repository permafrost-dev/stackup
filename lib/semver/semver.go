package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Semver struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Build      string
	String     string
}

// The function `ParseSemverString` takes a version string and returns a `Semver` struct with the
// parsed version components. The version string is expected to be in the format `major.minor.patch`,
// however this function will extract the version number from a string that contains other text as well.
func ParseSemverString(version string) *Semver {
	tempVersion, err := CoerceSemverString(ExtractVersion(version))
	if err != nil {
		return &Semver{}
	}

	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)(?:-(.+))?(?:\+(.+))?$`)
	matches := re.FindStringSubmatch(tempVersion)

	if len(matches) < 4 {
		return &Semver{}
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	prerelease := ""
	build := ""

	if len(matches) > 4 {
		prerelease = matches[4]
	}

	if len(matches) > 5 {
		build = matches[5]
	}

	return &Semver{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: prerelease,
		Build:      build,
		String:     tempVersion,
	}
}

func CoerceSemverString(version string) (string, error) {
	semverRegex := regexp.MustCompile(`^(?:[\^v~>]?)?(\d+)\.(\d+)\.(\d+)(?:-(.+))?(?:\+(.+))?$`)

	if semverRegex.MatchString(version) {
		return version, nil
	}

	semverRegex = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)(?:-(.+))?(?:\+(.+))`)
	if semverRegex.MatchString(version) {
		matches := semverRegex.FindAllString(version, -1)
		versionParts := matches[1:4]
		otherParts := matches[4:]
		return strings.Join([]string{strings.Join(versionParts, "."), strings.Join(otherParts, ".")}, "-"), nil
	}

	// If the input string does not match the semver regex, try to coerce it
	coercedRegex := regexp.MustCompile(`(\d+)\.(\d+)`)
	if coercedRegex.MatchString(version) {
		return fmt.Sprintf("%s.0", version), nil
	}

	coercedRegex = regexp.MustCompile(`(\d+)`)
	if coercedRegex.MatchString(version) {
		return fmt.Sprintf("%s.0.0", version), nil
	}

	// If the input string cannot be coerced into a semver string, return an error
	return "", fmt.Errorf("invalid semver string: %s", version)
}

// The function ExtractVersion extracts a version number from a given string using regular expressions.
func ExtractVersion(output string) string {
	versionRegex := regexp.MustCompile(`(\d+(\.\d+)?(\.\d+)?(\-.+$)?)`)

	match := versionRegex.FindStringSubmatch(output)
	if len(match) == 0 {
		return "0.0.0"
	}

	return strings.TrimSpace(match[0])
}

// The `Compare` method is a comparison method for the `Semver` struct. It compares the current
// `Semver` object with a given `version` string and returns an integer value indicating the result of
// the comparison. It returns -1 if the current `Semver` object is less than the `version` string, 0
// if the current `Semver` object is equal to the `version` string, and 1 if the current `Semver`
// is greater than the `version` string.
func (s *Semver) Compare(version string) int {
	semver1 := s
	semver2 := ParseSemverString(version)

	if semver1.Major < semver2.Major {
		return -1
	} else if semver1.Major > semver2.Major {
		return 1
	}

	if semver1.Minor < semver2.Minor {
		return -1
	} else if semver1.Minor > semver2.Minor {
		return 1
	}

	if semver1.Patch < semver2.Patch {
		return -1
	} else if semver1.Patch > semver2.Patch {
		return 1
	}

	if semver1.PreRelease == "" && semver2.PreRelease != "" {
		return 1
	} else if semver1.PreRelease != "" && semver2.PreRelease == "" {
		return -1
	} else if semver1.PreRelease != "" && semver2.PreRelease != "" {
		if semver1.PreRelease < semver2.PreRelease {
			return -1
		} else if semver1.PreRelease > semver2.PreRelease {
			return 1
		}
	}

	return 0
}

// The `GreaterThan` method is a comparison method for the `Semver` struct. It checks if the current
// `Semver` object is greater than the `otherVersion` string.
func (s *Semver) GreaterThan(otherVersion string) bool {
	other := ParseSemverString(otherVersion)

	if s.Major > other.Major {
		return true
	} else if s.Major < other.Major {
		return false
	}

	if s.Minor > other.Minor {
		return true
	} else if s.Minor < other.Minor {
		return false
	}

	if s.Patch > other.Patch {
		return true
	} else if s.Patch < other.Patch {
		return false
	}

	if s.PreRelease == "" && other.PreRelease != "" {
		return false
	} else if s.PreRelease != "" && other.PreRelease == "" {
		return true
	}

	return false
}

// The `Gte` method is a comparison method for the `Semver` struct. It checks if the current `Semver`
// object is greater than or equal to the `otherVersion` string.
func (s *Semver) Gte(otherVersion string) bool {
	return s.GreaterThan(otherVersion) || s.Equals(otherVersion)
}

// The `Lte` method is a comparison method for the `Semver` struct. It checks if the current `Semver`
// object is less than or equal to the `otherVersion` string.
func (s *Semver) Lte(otherVersion string) bool {
	return s.LessThan(otherVersion) || s.Equals(otherVersion)
}

// The `LessThan` method is a comparison method for the `Semver` struct. It checks if the current
// `Semver` object is less than the `otherVersion` string.
func (s *Semver) LessThan(otherVersion string) bool {
	other := ParseSemverString(otherVersion)

	if s.Major < other.Major {
		return true
	} else if s.Major > other.Major {
		return false
	}

	if s.Minor < other.Minor {
		return true
	} else if s.Minor > other.Minor {
		return false
	}

	if s.Patch < other.Patch {
		return true
	} else if s.Patch > other.Patch {
		return false
	}

	if s.PreRelease == "" && other.PreRelease != "" {
		return true
	} else if s.PreRelease != "" && other.PreRelease == "" {
		return false
	}

	return false
}

// The `Equals` method is a comparison method for the `Semver` struct. It checks if the current
// `Semver` object is equal to the `otherVersion` string.
func (s *Semver) Equals(otherVersion string) bool {
	other := ParseSemverString(otherVersion)

	return s.Major == other.Major &&
		s.Minor == other.Minor &&
		s.Patch == other.Patch &&
		s.PreRelease == other.PreRelease &&
		s.Build == other.Build
}
