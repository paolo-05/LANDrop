package update

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	PublishedAt string `json:"published_at"`
	HTMLURL     string `json:"html_url"`
}

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
}

// UpdateChecker handles checking for application updates
type UpdateChecker struct {
	repoOwner      string
	repoName       string
	currentVersion string
	app            fyne.App
}

// NewUpdateChecker creates a new update checker
func NewUpdateChecker(repoOwner, repoName, currentVersion string, app fyne.App) *UpdateChecker {
	return &UpdateChecker{
		repoOwner:      repoOwner,
		repoName:       repoName,
		currentVersion: currentVersion,
		app:            app,
	}
}

// ParseVersion parses a semantic version string (e.g., "v2.1.0" or "2.1.0")
func ParseVersion(versionStr string) (Version, error) {
	// Remove 'v' prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")

	// Use regex to extract major.minor.patch
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(versionStr)

	if len(matches) != 4 {
		return Version{}, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", matches[1])
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %s", matches[2])
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %s", matches[3])
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// ShouldUpdate determines if an update should be prompted based on version comparison
// Only major and minor version changes require updates, not patches
func ShouldUpdate(current, latest Version) bool {
	// Major version change
	if latest.Major > current.Major {
		return true
	}

	// Minor version change (within the same major version)
	if latest.Major == current.Major && latest.Minor > current.Minor {
		return true
	}

	// Patch changes don't require updates
	return false
}

// GetLastUpdateCheck returns the timestamp of the last update check
func (uc *UpdateChecker) GetLastUpdateCheck() time.Time {
	timestamp := uc.app.Preferences().Int("last_update_check")
	if timestamp == 0 {
		return time.Time{} // Zero time if never checked
	}
	return time.Unix(int64(timestamp), 0)
}

// SetLastUpdateCheck stores the timestamp of the last update check
func (uc *UpdateChecker) SetLastUpdateCheck(t time.Time) {
	uc.app.Preferences().SetInt("last_update_check", int(t.Unix()))
}

// GetSkippedVersion returns the version that the user chose to skip
func (uc *UpdateChecker) GetSkippedVersion() string {
	return uc.app.Preferences().String("skipped_version")
}

// SetSkippedVersion stores the version that the user chose to skip
func (uc *UpdateChecker) SetSkippedVersion(version string) {
	uc.app.Preferences().SetString("skipped_version", version)
}

// IsUpdateCheckEnabled returns whether automatic update checking is enabled
func (uc *UpdateChecker) IsUpdateCheckEnabled() bool {
	return uc.app.Preferences().BoolWithFallback("auto_update_check", true)
}

// SetUpdateCheckEnabled sets whether automatic update checking is enabled
func (uc *UpdateChecker) SetUpdateCheckEnabled(enabled bool) {
	uc.app.Preferences().SetBool("auto_update_check", enabled)
}

// ShouldCheckForUpdates determines if we should check for updates based on timing
func (uc *UpdateChecker) ShouldCheckForUpdates() bool {
	if !uc.IsUpdateCheckEnabled() {
		return false
	}

	lastCheck := uc.GetLastUpdateCheck()
	if lastCheck.IsZero() {
		return true // Never checked before
	}

	// Check once per day
	return time.Since(lastCheck) > 24*time.Hour
}

// FetchLatestRelease fetches the latest release from GitHub
func (uc *UpdateChecker) FetchLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", uc.repoOwner, uc.repoName)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release JSON: %w", err)
	}

	return &release, nil
}

// CheckForUpdates checks for updates and returns update information
func (uc *UpdateChecker) CheckForUpdates() (*UpdateInfo, error) {
	// Parse current version
	currentVer, err := ParseVersion(uc.currentVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid current version: %w", err)
	}

	// Fetch latest release
	release, err := uc.FetchLatestRelease()
	if err != nil {
		return nil, err
	}

	// Skip draft or prerelease versions
	if release.Draft || release.Prerelease {
		log.Printf("Skipping draft/prerelease version: %s", release.TagName)
		return nil, nil
	}

	// Parse latest version
	latestVer, err := ParseVersion(release.TagName)
	if err != nil {
		return nil, fmt.Errorf("invalid latest version: %w", err)
	}

	// Check if user already skipped this version
	skippedVersion := uc.GetSkippedVersion()
	if skippedVersion == release.TagName {
		log.Printf("User already skipped version: %s", release.TagName)
		return nil, nil
	}

	// Determine if update should be prompted
	if ShouldUpdate(currentVer, latestVer) {
		return &UpdateInfo{
			Available:      true,
			CurrentVersion: uc.currentVersion,
			LatestVersion:  release.TagName,
			ReleaseNotes:   release.Body,
			DownloadURL:    release.HTMLURL,
			IsMinorUpdate:  latestVer.Major == currentVer.Major,
		}, nil
	}

	log.Printf("No significant update available. Current: %s, Latest: %s", uc.currentVersion, release.TagName)
	return nil, nil
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseNotes   string
	DownloadURL    string
	IsMinorUpdate  bool // true if it's a minor update, false if major
}

// CheckForUpdatesAsync checks for updates in the background
func (uc *UpdateChecker) CheckForUpdatesAsync(callback func(*UpdateInfo, error)) {
	go func() {
		// Record that we checked for updates
		uc.SetLastUpdateCheck(time.Now())

		updateInfo, err := uc.CheckForUpdates()
		callback(updateInfo, err)
	}()
}
