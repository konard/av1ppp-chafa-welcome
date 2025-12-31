package share

import (
	"testing"
)

func TestExtractImagePath(t *testing.T) {
	tests := []struct {
		config   string
		expected string
	}{
		{
			config:   "source = '/path/to/image.jpg'",
			expected: "/path/to/image.jpg",
		},
		{
			config:   `source = "/path/to/image.png"`,
			expected: "/path/to/image.png",
		},
		{
			config:   "[image]\nsource = '/my/image.jpg'\nsize = 32",
			expected: "/my/image.jpg",
		},
		{
			config:   "no source here",
			expected: "",
		},
		{
			config:   "  source='/with/spaces.jpg'",
			expected: "/with/spaces.jpg",
		},
	}

	for _, tt := range tests {
		result := extractImagePath(tt.config)
		if result != tt.expected {
			t.Errorf("extractImagePath(%q) = %q, want %q", tt.config, result, tt.expected)
		}
	}
}

func TestExtractGistID(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{
			url:      "https://gist.github.com/username/abc123def456",
			expected: "abc123def456",
		},
		{
			url:      "https://gist.github.com/abc123def456",
			expected: "abc123def456",
		},
		{
			url:      "abc123def456",
			expected: "abc123def456",
		},
		{
			url:      "  abc123def456  ",
			expected: "abc123def456",
		},
		{
			url:      "https://gist.github.com/user/abc123/file",
			expected: "file",
		},
	}

	for _, tt := range tests {
		result := extractGistID(tt.url)
		if result != tt.expected {
			t.Errorf("extractGistID(%q) = %q, want %q", tt.url, result, tt.expected)
		}
	}
}

func TestShareOptionsDefaults(t *testing.T) {
	opts := ShareOptions{}

	if opts.IncludeImage {
		t.Error("Expected IncludeImage to be false by default")
	}

	if opts.Public {
		t.Error("Expected Public to be false by default")
	}

	if opts.UseGhCli {
		t.Error("Expected UseGhCli to be false by default")
	}
}

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("Expected non-nil Share instance")
	}

	if s.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
}
