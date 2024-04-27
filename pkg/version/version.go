package version

import "github.com/Masterminds/semver/v3"

var (
	Version    = "0.1.0"
	CommitHash = "N/A"
)

func IsCompatible(myVersion string, otherVersion string) bool {
	mySemver, err := semver.NewVersion(myVersion)

	// If one of the versions cannot be parsed, mark them as incompatible
	if err != nil {
		return false
	}

	otherSemver, err := semver.NewVersion(otherVersion)
	if err != nil {
		return false
	}

	if mySemver.Equal(otherSemver) {
		return true
	}

	return false
}
