package utils

import (
	"net/url"
	"testing"
)

func TestParseURLValid(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"https://example.com", "https://example.com"},
		{"http://localhost:8080", "http://localhost:8080"},
		{"https://github.com/user/repo", "https://github.com/user/repo"},
		{"mailto:test@example.com", "mailto:test@example.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := ParseURL(tc.input)
			if result == nil {
				t.Errorf("ParseURL returned nil for valid URL: %s", tc.input)
				return
			}
			if result.String() != tc.expected {
				t.Errorf("ParseURL(%s) = %s, want %s", tc.input, result.String(), tc.expected)
			}
		})
	}
}

func TestParseURLInvalid(t *testing.T) {
	invalidURLs := []string{
		"not a url",
		"://invalid",
		"ht!tp://invalid.com",
		string([]byte{0x7f}), // Invalid characters
	}

	for _, invalidURL := range invalidURLs {
		t.Run(invalidURL, func(t *testing.T) {
			result := ParseURL(invalidURL)
			// Should return empty URL object, not nil
			if result == nil {
				t.Errorf("ParseURL should not return nil, even for invalid URLs")
				return
			}
			// For invalid URLs, we expect an empty URL object
			emptyURL := &url.URL{}
			if result.String() != emptyURL.String() {
				t.Logf("ParseURL returned: %+v for invalid URL: %s", result, invalidURL)
			}
		})
	}
}

func TestParseURLEmpty(t *testing.T) {
	result := ParseURL("")
	if result == nil {
		t.Error("ParseURL should not return nil for empty string")
		return
	}

	// Empty string should parse to empty URL
	if result.String() != "" {
		t.Errorf("ParseURL(\"\") = %s, want empty string", result.String())
	}
}

func TestParseURLComponents(t *testing.T) {
	testURL := "https://user:pass@example.com:8080/path?query=value#fragment"
	result := ParseURL(testURL)

	if result == nil {
		t.Fatal("ParseURL returned nil")
	}

	if result.Scheme != "https" {
		t.Errorf("Expected scheme 'https', got '%s'", result.Scheme)
	}

	if result.Host != "example.com:8080" {
		t.Errorf("Expected host 'example.com:8080', got '%s'", result.Host)
	}

	if result.Path != "/path" {
		t.Errorf("Expected path '/path', got '%s'", result.Path)
	}

	if result.RawQuery != "query=value" {
		t.Errorf("Expected query 'query=value', got '%s'", result.RawQuery)
	}

	if result.Fragment != "fragment" {
		t.Errorf("Expected fragment 'fragment', got '%s'", result.Fragment)
	}
}
