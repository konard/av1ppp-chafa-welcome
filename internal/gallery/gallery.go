package gallery

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/av1ppp/chafa-welcome/internal/global"
	"github.com/fatih/color"
)

const (
	// DefaultGalleryURL is the URL to fetch community presets from
	DefaultGalleryURL = "https://raw.githubusercontent.com/av1ppp/chafa-welcome-presets/main/index.json"
	// CacheFileName is the name of the cached gallery index
	CacheFileName = "gallery_cache.json"
	// CacheDuration is how long to cache the gallery index
	CacheDuration = 1 * time.Hour
)

// Gallery handles browsing and applying community presets
type Gallery struct {
	url       string
	cacheDir  string
	presets   []Preset
	httpClient *http.Client
}

// New creates a new Gallery instance
func New() *Gallery {
	return &Gallery{
		url:       DefaultGalleryURL,
		cacheDir:  filepath.Join(global.HomeDir(), "gallery"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewWithURL creates a new Gallery with a custom URL
func NewWithURL(url string) *Gallery {
	g := New()
	g.url = url
	return g
}

// Run starts the interactive gallery browser
func (g *Gallery) Run() error {
	// Ensure gallery cache directory exists
	if err := os.MkdirAll(g.cacheDir, global.ModeDir); err != nil {
		return fmt.Errorf("failed to create gallery cache directory: %w", err)
	}

	// Load presets (from cache or network)
	if err := g.loadPresets(); err != nil {
		return fmt.Errorf("failed to load presets: %w", err)
	}

	if len(g.presets) == 0 {
		fmt.Println("No presets available in the gallery yet.")
		fmt.Println("Check back later or contribute your own at:")
		fmt.Println("  https://github.com/av1ppp/chafa-welcome-presets")
		return nil
	}

	// Start interactive browser
	return g.browse()
}

// loadPresets loads presets from cache or fetches from network
func (g *Gallery) loadPresets() error {
	cachePath := filepath.Join(g.cacheDir, CacheFileName)

	// Check if cache exists and is fresh
	if info, err := os.Stat(cachePath); err == nil {
		if time.Since(info.ModTime()) < CacheDuration {
			// Use cached data
			data, err := os.ReadFile(cachePath)
			if err == nil {
				var index PresetIndex
				if err := json.Unmarshal(data, &index); err == nil {
					g.presets = index.Presets
					return nil
				}
			}
		}
	}

	// Fetch from network
	return g.fetchPresets()
}

// fetchPresets fetches presets from the remote gallery
func (g *Gallery) fetchPresets() error {
	fmt.Println("Fetching community presets...")

	resp, err := g.httpClient.Get(g.url)
	if err != nil {
		return fmt.Errorf("failed to fetch gallery: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If gallery doesn't exist yet, return empty presets
		if resp.StatusCode == http.StatusNotFound {
			g.presets = []Preset{}
			return nil
		}
		return fmt.Errorf("gallery server returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read gallery response: %w", err)
	}

	var index PresetIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("failed to parse gallery index: %w", err)
	}

	g.presets = index.Presets

	// Cache the data
	cachePath := filepath.Join(g.cacheDir, CacheFileName)
	_ = os.WriteFile(cachePath, data, global.ModeFile)

	return nil
}

// browse starts the interactive preset browser
func (g *Gallery) browse() error {
	reader := bufio.NewReader(os.Stdin)
	sortBy := SortByPopular
	categoryFilter := ""

	for {
		g.displayPresets(sortBy, categoryFilter)

		fmt.Println()
		bold := color.New(color.Bold)
		bold.Println("Commands:")
		fmt.Println("  [number] - Apply preset")
		fmt.Println("  s - Sort by (popular/newest/name)")
		fmt.Println("  c - Filter by category")
		fmt.Println("  r - Refresh from server")
		fmt.Println("  q - Quit")
		fmt.Println()
		fmt.Print("Enter command: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "q", "quit", "exit":
			fmt.Println("Goodbye!")
			return nil
		case "s", "sort":
			sortBy = g.promptSort(reader)
		case "c", "category":
			categoryFilter = g.promptCategory(reader)
		case "r", "refresh":
			if err := g.fetchPresets(); err != nil {
				fmt.Printf("Error refreshing: %v\n", err)
			}
		default:
			// Try to parse as a number
			if num, err := strconv.Atoi(input); err == nil {
				if err := g.applyPreset(num, reader); err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			} else {
				fmt.Println("Unknown command. Try 'q' to quit.")
			}
		}
	}
}

// displayPresets shows the list of presets
func (g *Gallery) displayPresets(sortBy SortOption, categoryFilter string) {
	// Clear screen (simple approach)
	fmt.Print("\033[H\033[2J")

	titleColor := color.New(color.FgCyan, color.Bold)
	titleColor.Println("=== Chafa-Welcome Community Gallery ===")
	fmt.Println()

	// Filter and sort presets
	filtered := g.filterPresets(categoryFilter)
	g.sortPresets(filtered, sortBy)

	if len(filtered) == 0 {
		fmt.Println("No presets match your criteria.")
		return
	}

	headerColor := color.New(color.FgBlue, color.Bold)
	headerColor.Printf("%-4s %-25s %-15s %-15s %s\n", "#", "Name", "Author", "Category", "Downloads")
	fmt.Println(strings.Repeat("-", 70))

	for i, p := range filtered {
		fmt.Printf("%-4d %-25s %-15s %-15s %d\n",
			i+1,
			truncate(p.Name, 25),
			truncate(p.Author, 15),
			truncate(p.Category, 15),
			p.Downloads,
		)
	}
}

// filterPresets filters presets by category
func (g *Gallery) filterPresets(category string) []Preset {
	if category == "" {
		return g.presets
	}

	var filtered []Preset
	for _, p := range g.presets {
		if strings.EqualFold(p.Category, category) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// sortPresets sorts presets by the given option
func (g *Gallery) sortPresets(presets []Preset, sortBy SortOption) {
	switch sortBy {
	case SortByPopular:
		sort.Slice(presets, func(i, j int) bool {
			return presets[i].Downloads > presets[j].Downloads
		})
	case SortByNewest:
		// For now, just reverse order (assuming newer entries are added at the end)
		for i, j := 0, len(presets)-1; i < j; i, j = i+1, j-1 {
			presets[i], presets[j] = presets[j], presets[i]
		}
	case SortByName:
		sort.Slice(presets, func(i, j int) bool {
			return strings.ToLower(presets[i].Name) < strings.ToLower(presets[j].Name)
		})
	}
}

// promptSort prompts user for sort option
func (g *Gallery) promptSort(reader *bufio.Reader) SortOption {
	fmt.Println()
	fmt.Println("Sort by:")
	fmt.Println("  1 - Popular (default)")
	fmt.Println("  2 - Newest")
	fmt.Println("  3 - Name")
	fmt.Print("Choose: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "2":
		return SortByNewest
	case "3":
		return SortByName
	default:
		return SortByPopular
	}
}

// promptCategory prompts user for category filter
func (g *Gallery) promptCategory(reader *bufio.Reader) string {
	categories := g.getCategories()

	fmt.Println()
	fmt.Println("Categories:")
	fmt.Println("  0 - All (no filter)")
	for i, cat := range categories {
		fmt.Printf("  %d - %s (%d presets)\n", i+1, cat.Name, cat.Count)
	}
	fmt.Print("Choose: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if num, err := strconv.Atoi(input); err == nil {
		if num == 0 {
			return ""
		}
		if num > 0 && num <= len(categories) {
			return categories[num-1].Name
		}
	}

	return ""
}

// getCategories returns list of categories with counts
func (g *Gallery) getCategories() []Category {
	counts := make(map[string]int)
	for _, p := range g.presets {
		if p.Category != "" {
			counts[p.Category]++
		}
	}

	var categories []Category
	for name, count := range counts {
		categories = append(categories, Category{Name: name, Count: count})
	}

	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Count > categories[j].Count
	})

	return categories
}

// applyPreset applies a preset by number
func (g *Gallery) applyPreset(num int, reader *bufio.Reader) error {
	if num < 1 || num > len(g.presets) {
		return fmt.Errorf("invalid preset number: %d", num)
	}

	preset := g.presets[num-1]

	// Show preset details
	fmt.Println()
	detailColor := color.New(color.FgYellow, color.Bold)
	detailColor.Printf("Preset: %s\n", preset.Name)
	fmt.Printf("Author: %s\n", preset.Author)
	fmt.Printf("Category: %s\n", preset.Category)
	fmt.Printf("Description: %s\n", preset.Description)
	fmt.Println()
	fmt.Print("Apply this preset? (y/n): ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	return g.downloadAndApply(preset)
}

// downloadAndApply downloads and applies a preset
func (g *Gallery) downloadAndApply(preset Preset) error {
	fmt.Println("Downloading preset...")

	// Download config
	resp, err := g.httpClient.Get(preset.ConfigURL)
	if err != nil {
		return fmt.Errorf("failed to download config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("config download returned status %d", resp.StatusCode)
	}

	configData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Download image if available
	var imagePath string
	if preset.ImageURL != "" {
		imagePath, err = g.downloadImage(preset)
		if err != nil {
			fmt.Printf("Warning: failed to download image: %v\n", err)
			fmt.Println("You may need to update the image path in the config manually.")
		}
	}

	// Apply config
	configPath := filepath.Join(global.HomeDir(), "config")

	// Backup existing config
	backupPath := configPath + ".backup"
	if _, err := os.Stat(configPath); err == nil {
		data, _ := os.ReadFile(configPath)
		_ = os.WriteFile(backupPath, data, global.ModeFile)
		fmt.Printf("Backed up existing config to: %s\n", backupPath)
	}

	// Update image path in config if we downloaded an image
	if imagePath != "" {
		configStr := string(configData)
		// Simple replacement - look for source field and update it
		configStr = updateImageSource(configStr, imagePath)
		configData = []byte(configStr)
	}

	if err := os.WriteFile(configPath, configData, global.ModeFile); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	successColor := color.New(color.FgGreen, color.Bold)
	successColor.Println("Preset applied successfully!")
	fmt.Println("Run 'chafa-welcome' to see your new configuration.")

	return nil
}

// downloadImage downloads the preset's image
func (g *Gallery) downloadImage(preset Preset) (string, error) {
	resp, err := g.httpClient.Get(preset.ImageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("image download returned status %d", resp.StatusCode)
	}

	// Determine file extension
	ext := filepath.Ext(preset.ImageURL)
	if ext == "" {
		ext = ".jpg"
	}

	// Save to images directory
	imagesDir := filepath.Join(global.HomeDir(), "images")
	if err := os.MkdirAll(imagesDir, global.ModeDir); err != nil {
		return "", err
	}

	// Use preset name as filename (sanitized)
	safeName := strings.ReplaceAll(preset.Name, " ", "_")
	safeName = strings.ReplaceAll(safeName, "/", "_")
	imagePath := filepath.Join(imagesDir, safeName+ext)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(imagePath, data, global.ModeFile); err != nil {
		return "", err
	}

	return imagePath, nil
}

// updateImageSource updates the image source in a TOML config string
func updateImageSource(config, imagePath string) string {
	lines := strings.Split(config, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "source") && strings.Contains(trimmed, "=") {
			// Replace the source line
			lines[i] = fmt.Sprintf("source = '%s'", imagePath)
			break
		}
	}
	return strings.Join(lines, "\n")
}

// truncate truncates a string to the given length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
