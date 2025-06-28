package update

import (
	"fmt"

	"github.com/blang/semver"
)

const (
	versionInfoURL = "" // Loads version.json File From CDN - I Used Digital Oceans
	downloadURL    = "" // Download New Build From CDN - I Used Digital Oceans
)

func CheckAndUpdate(currentVersion string) (bool, error) {
	latestVersion, err := fetchLatestVersion()
	if err != nil {
		return false, fmt.Errorf("failed to fetch latest version: %w", err)
	}

	currentSemver, err := semver.Parse(currentVersion)
	if err != nil {
		return false, fmt.Errorf("invalid current version: %w", err)
	}

	latestSemver, err := semver.Parse(latestVersion)
	if err != nil {
		return false, fmt.Errorf("invalid latest version: %w", err)
	}

	if latestSemver.GT(currentSemver) {
		return true, nil
	}
	return false, nil
}
