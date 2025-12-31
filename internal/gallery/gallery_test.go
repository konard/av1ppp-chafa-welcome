package gallery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPresetIndex(t *testing.T) {
	index := PresetIndex{
		Version:    "1.0",
		LastUpdate: "2024-01-01",
		Presets: []Preset{
			{
				Name:        "Cyberpunk",
				Description: "A cyberpunk themed config",
				Author:      "test_user",
				Category:    "Cyberpunk",
				Downloads:   100,
				ConfigURL:   "https://example.com/config.toml",
			},
		},
	}

	data, err := json.Marshal(index)
	if err != nil {
		t.Fatalf("Failed to marshal preset index: %v", err)
	}

	var decoded PresetIndex
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal preset index: %v", err)
	}

	if decoded.Version != index.Version {
		t.Errorf("Expected version %s, got %s", index.Version, decoded.Version)
	}

	if len(decoded.Presets) != 1 {
		t.Fatalf("Expected 1 preset, got %d", len(decoded.Presets))
	}

	if decoded.Presets[0].Name != "Cyberpunk" {
		t.Errorf("Expected preset name 'Cyberpunk', got '%s'", decoded.Presets[0].Name)
	}
}

func TestSortOptionString(t *testing.T) {
	tests := []struct {
		opt      SortOption
		expected string
	}{
		{SortByPopular, "popular"},
		{SortByNewest, "newest"},
		{SortByName, "name"},
		{SortOption(99), "popular"}, // Unknown should default to popular
	}

	for _, tt := range tests {
		result := tt.opt.String()
		if result != tt.expected {
			t.Errorf("SortOption(%d).String() = %s, want %s", tt.opt, result, tt.expected)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
		{"abcd", 6, "abcd"}, // len(4) <= 6, no truncation needed
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestUpdateImageSource(t *testing.T) {
	tests := []struct {
		config    string
		imagePath string
		expected  string
	}{
		{
			config:    "source = '/old/path.jpg'",
			imagePath: "/new/path.jpg",
			expected:  "source = '/new/path.jpg'",
		},
		{
			config:    "[image]\nsource = '/old/path.jpg'\nsize = 32",
			imagePath: "/new/path.jpg",
			expected:  "[image]\nsource = '/new/path.jpg'\nsize = 32",
		},
		{
			config:    "no source here",
			imagePath: "/new/path.jpg",
			expected:  "no source here",
		},
	}

	for _, tt := range tests {
		result := updateImageSource(tt.config, tt.imagePath)
		if result != tt.expected {
			t.Errorf("updateImageSource() = %q, want %q", result, tt.expected)
		}
	}
}

func TestFilterPresets(t *testing.T) {
	g := &Gallery{
		presets: []Preset{
			{Name: "Preset1", Category: "Cyberpunk"},
			{Name: "Preset2", Category: "Minimal"},
			{Name: "Preset3", Category: "Cyberpunk"},
			{Name: "Preset4", Category: "Retro"},
		},
	}

	// Test filter by category
	filtered := g.filterPresets("Cyberpunk")
	if len(filtered) != 2 {
		t.Errorf("Expected 2 presets with category Cyberpunk, got %d", len(filtered))
	}

	// Test no filter
	all := g.filterPresets("")
	if len(all) != 4 {
		t.Errorf("Expected 4 presets with no filter, got %d", len(all))
	}

	// Test non-existent category
	none := g.filterPresets("NonExistent")
	if len(none) != 0 {
		t.Errorf("Expected 0 presets for non-existent category, got %d", len(none))
	}
}

func TestSortPresets(t *testing.T) {
	g := &Gallery{}

	// Test sort by name
	presets := []Preset{
		{Name: "Zebra", Downloads: 10},
		{Name: "Alpha", Downloads: 50},
		{Name: "Beta", Downloads: 30},
	}

	g.sortPresets(presets, SortByName)
	if presets[0].Name != "Alpha" {
		t.Errorf("Expected first preset to be Alpha after name sort, got %s", presets[0].Name)
	}

	// Test sort by popular
	presets = []Preset{
		{Name: "Low", Downloads: 10},
		{Name: "High", Downloads: 100},
		{Name: "Mid", Downloads: 50},
	}

	g.sortPresets(presets, SortByPopular)
	if presets[0].Downloads != 100 {
		t.Errorf("Expected first preset to have 100 downloads after popular sort, got %d", presets[0].Downloads)
	}
}

func TestGetCategories(t *testing.T) {
	g := &Gallery{
		presets: []Preset{
			{Category: "Cyberpunk"},
			{Category: "Cyberpunk"},
			{Category: "Minimal"},
			{Category: "Cyberpunk"},
			{Category: "Retro"},
		},
	}

	categories := g.getCategories()

	if len(categories) != 3 {
		t.Fatalf("Expected 3 categories, got %d", len(categories))
	}

	// Should be sorted by count (descending)
	if categories[0].Name != "Cyberpunk" || categories[0].Count != 3 {
		t.Errorf("Expected first category to be Cyberpunk with 3, got %s with %d",
			categories[0].Name, categories[0].Count)
	}
}

func TestFetchPresetsNotFound(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	g := NewWithURL(server.URL)
	g.cacheDir = t.TempDir()

	err := g.fetchPresets()
	if err != nil {
		t.Errorf("Expected no error for 404 (empty gallery), got %v", err)
	}

	if len(g.presets) != 0 {
		t.Errorf("Expected empty presets for 404, got %d", len(g.presets))
	}
}

func TestFetchPresetsSuccess(t *testing.T) {
	index := PresetIndex{
		Version:    "1.0",
		LastUpdate: "2024-01-01",
		Presets: []Preset{
			{Name: "Test", Author: "tester", Category: "Test"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(index)
	}))
	defer server.Close()

	g := NewWithURL(server.URL)
	g.cacheDir = t.TempDir()

	err := g.fetchPresets()
	if err != nil {
		t.Fatalf("Failed to fetch presets: %v", err)
	}

	if len(g.presets) != 1 {
		t.Fatalf("Expected 1 preset, got %d", len(g.presets))
	}

	if g.presets[0].Name != "Test" {
		t.Errorf("Expected preset name 'Test', got '%s'", g.presets[0].Name)
	}
}
