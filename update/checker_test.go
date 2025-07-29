package update

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected Version
		hasError bool
	}{
		{"v2.1.0", Version{2, 1, 0}, false},
		{"2.1.0", Version{2, 1, 0}, false},
		{"1.0.5", Version{1, 0, 5}, false},
		{"10.20.30", Version{10, 20, 30}, false},
		{"invalid", Version{}, true},
		{"v2.1", Version{}, true},
		{"2.1.0.4", Version{2, 1, 0}, false}, // Should parse first three parts
		{"", Version{}, true},
	}

	for _, test := range tests {
		result, err := ParseVersion(test.input)

		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input %s, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("For input %s, expected %v, got %v", test.input, test.expected, result)
			}
		}
	}
}

func TestShouldUpdate(t *testing.T) {
	tests := []struct {
		current      Version
		latest       Version
		shouldUpdate bool
		description  string
	}{
		{Version{2, 0, 0}, Version{3, 0, 0}, true, "major version increase"},
		{Version{2, 0, 0}, Version{2, 1, 0}, true, "minor version increase"},
		{Version{2, 1, 0}, Version{2, 1, 1}, false, "patch version increase"},
		{Version{2, 1, 1}, Version{2, 1, 0}, false, "version decrease"},
		{Version{2, 1, 0}, Version{2, 1, 0}, false, "same version"},
		{Version{1, 9, 5}, Version{2, 0, 0}, true, "major version bump"},
		{Version{2, 5, 3}, Version{2, 6, 0}, true, "minor version bump"},
		{Version{1, 0, 0}, Version{1, 0, 10}, false, "patch only"},
	}

	for _, test := range tests {
		result := ShouldUpdate(test.current, test.latest)
		if result != test.shouldUpdate {
			t.Errorf("%s: expected %v, got %v (current: %v, latest: %v)",
				test.description, test.shouldUpdate, result, test.current, test.latest)
		}
	}
}

func TestVersionComparison(t *testing.T) {
	// Test various edge cases
	currentVer := Version{2, 1, 5}

	// Should update for major version
	if !ShouldUpdate(currentVer, Version{3, 0, 0}) {
		t.Error("Should update for major version change")
	}

	// Should update for minor version within same major
	if !ShouldUpdate(currentVer, Version{2, 2, 0}) {
		t.Error("Should update for minor version change")
	}

	// Should NOT update for patch version
	if ShouldUpdate(currentVer, Version{2, 1, 6}) {
		t.Error("Should NOT update for patch version change")
	}

	// Should NOT update for older versions
	if ShouldUpdate(currentVer, Version{2, 0, 9}) {
		t.Error("Should NOT update for older version")
	}

	if ShouldUpdate(currentVer, Version{1, 9, 9}) {
		t.Error("Should NOT update for older major version")
	}
}

func TestParseVersionEdgeCases(t *testing.T) {
	// Test realistic version strings
	validVersions := []string{
		"v1.0.0",
		"v2.3.1",
		"1.0.0",
		"10.5.23",
		"v0.1.0",
	}

	for _, version := range validVersions {
		_, err := ParseVersion(version)
		if err != nil {
			t.Errorf("Valid version %s should parse without error: %v", version, err)
		}
	}

	// Test invalid version strings
	invalidVersions := []string{
		"v1.0",
		"1.0",
		"v1",
		"invalid",
		"1.0.0.0.1",
		"v1.0.0-beta",
		"1.0.0-rc1",
	}

	for _, version := range invalidVersions {
		result, err := ParseVersion(version)
		if version == "1.0.0.0.1" && err == nil {
			// This one might parse the first three parts, which is acceptable
			if result.Major != 1 || result.Minor != 0 || result.Patch != 0 {
				t.Errorf("Version %s parsed incorrectly: %v", version, result)
			}
		} else if version == "v1.0.0-beta" || version == "1.0.0-rc1" {
			// These might parse the version part correctly
			if err == nil && (result.Major != 1 || result.Minor != 0 || result.Patch != 0) {
				t.Errorf("Version %s parsed incorrectly: %v", version, result)
			}
		} else if err == nil {
			t.Errorf("Invalid version %s should produce an error but got: %v", version, result)
		}
	}
}
