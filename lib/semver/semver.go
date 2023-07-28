package semver

import (
	"regexp"
	"strconv"
)

type Semver struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Build      string
	String     string
}

func ParseSemverString(version string) Semver {
	// Match the major, minor, and patch version numbers
	re := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-(.+))?(?:\+(.+))?$`)
	matches := re.FindStringSubmatch(version)

	if len(matches) < 4 {
		return Semver{}
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

	return Semver{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: prerelease,
		Build:      build,
		String:     version,
	}
}

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

func (s *Semver) GreaterThan(other Semver) bool {
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

func (s *Semver) LessThan(other Semver) bool {
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

func (s *Semver) Equals(other Semver) bool {
	return s.Major == other.Major &&
		s.Minor == other.Minor &&
		s.Patch == other.Patch &&
		s.PreRelease == other.PreRelease &&
		s.Build == other.Build
}
