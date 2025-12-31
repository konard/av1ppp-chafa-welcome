package share

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/av1ppp/chafa-welcome/internal/global"
	"github.com/fatih/color"
)

// GistFile represents a file in a GitHub Gist
type GistFile struct {
	Content string `json:"content"`
}

// GistRequest represents the request body for creating a Gist
type GistRequest struct {
	Description string              `json:"description"`
	Public      bool                `json:"public"`
	Files       map[string]GistFile `json:"files"`
}

// GistResponse represents the response from creating a Gist
type GistResponse struct {
	HTMLURL string `json:"html_url"`
	ID      string `json:"id"`
}

// ShareOptions contains options for sharing
type ShareOptions struct {
	IncludeImage bool
	Public       bool
	UseGhCli     bool // Use gh CLI instead of API
}

// Share handles sharing the current configuration
type Share struct {
	httpClient *http.Client
}

// New creates a new Share instance
func New() *Share {
	return &Share{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Run shares the current configuration
func (s *Share) Run(opts ShareOptions) error {
	homeDir := global.HomeDir()
	configPath := filepath.Join(homeDir, "config")

	// Read config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config to get image path if needed
	var imageData []byte
	var imageName string
	if opts.IncludeImage {
		imagePath := extractImagePath(string(configData))
		if imagePath != "" {
			imageData, err = os.ReadFile(imagePath)
			if err != nil {
				fmt.Printf("Warning: could not read image file: %v\n", err)
				fmt.Println("Sharing config without image.")
			} else {
				imageName = filepath.Base(imagePath)
			}
		}
	}

	// Create gist
	if opts.UseGhCli {
		return s.shareWithGhCli(configData, imageData, imageName, opts.Public)
	}
	return s.shareWithAPI(configData, imageData, imageName, opts.Public)
}

// shareWithGhCli uses the GitHub CLI to create a gist
func (s *Share) shareWithGhCli(configData, imageData []byte, imageName string, public bool) error {
	// Check if gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("GitHub CLI (gh) not found. Install it from https://cli.github.com/ or set GITHUB_TOKEN environment variable")
	}

	// Create temporary directory for files
	tmpDir, err := os.MkdirTemp("", "chafa-welcome-share-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write config file
	configFile := filepath.Join(tmpDir, "chafa-welcome-config.toml")
	if err := os.WriteFile(configFile, configData, 0644); err != nil {
		return fmt.Errorf("failed to write temp config: %w", err)
	}

	// Build gh command
	args := []string{"gist", "create", configFile}
	if public {
		args = append(args, "--public")
	}
	args = append(args, "-d", "My chafa-welcome configuration")

	// Add image if available (base64 encoded)
	if len(imageData) > 0 && imageName != "" {
		imageFile := filepath.Join(tmpDir, imageName+".base64.txt")
		encoded := base64.StdEncoding.EncodeToString(imageData)
		header := fmt.Sprintf("# Base64 encoded image: %s\n# To decode: base64 -d %s.base64.txt > %s\n\n", imageName, imageName, imageName)
		if err := os.WriteFile(imageFile, []byte(header+encoded), 0644); err != nil {
			fmt.Printf("Warning: could not include image: %v\n", err)
		} else {
			args = append(args, imageFile)
		}
	}

	fmt.Println("Creating gist...")
	cmd := exec.Command("gh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create gist: %w\nOutput: %s", err, string(output))
	}

	url := strings.TrimSpace(string(output))
	s.printSuccess(url)
	return nil
}

// shareWithAPI uses the GitHub API directly to create a gist
func (s *Share) shareWithAPI(configData, imageData []byte, imageName string, public bool) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("No GITHUB_TOKEN found. Trying GitHub CLI...")
		return s.shareWithGhCli(configData, imageData, imageName, public)
	}

	files := map[string]GistFile{
		"chafa-welcome-config.toml": {Content: string(configData)},
	}

	// Add image as base64 if available
	if len(imageData) > 0 && imageName != "" {
		encoded := base64.StdEncoding.EncodeToString(imageData)
		header := fmt.Sprintf("# Base64 encoded image: %s\n# To decode: base64 -d %s.base64.txt > %s\n\n", imageName, imageName, imageName)
		files[imageName+".base64.txt"] = GistFile{Content: header + encoded}
	}

	gistReq := GistRequest{
		Description: "My chafa-welcome configuration",
		Public:      public,
		Files:       files,
	}

	jsonData, err := json.Marshal(gistReq)
	if err != nil {
		return fmt.Errorf("failed to marshal gist request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/gists", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	fmt.Println("Creating gist...")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create gist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create gist (status %d): %s", resp.StatusCode, string(body))
	}

	var gistResp GistResponse
	if err := json.NewDecoder(resp.Body).Decode(&gistResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	s.printSuccess(gistResp.HTMLURL)
	return nil
}

// printSuccess prints the success message with the gist URL
func (s *Share) printSuccess(url string) {
	fmt.Println()
	successColor := color.New(color.FgGreen, color.Bold)
	successColor.Println("Configuration shared successfully!")
	fmt.Println()

	linkColor := color.New(color.FgCyan, color.Underline)
	fmt.Print("Gist URL: ")
	linkColor.Println(url)
	fmt.Println()

	fmt.Println("Share this link with others!")
	fmt.Println()
	fmt.Println("To submit to the community gallery, create a Pull Request at:")
	fmt.Println("  https://github.com/av1ppp/chafa-welcome-presets")
}

// extractImagePath extracts the image source path from config content
func extractImagePath(config string) string {
	lines := strings.Split(config, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "source") && strings.Contains(trimmed, "=") {
			// Extract value between quotes
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				// Remove quotes
				value = strings.Trim(value, "'\"")
				return value
			}
		}
	}
	return ""
}

// Import imports a configuration from a gist URL
func (s *Share) Import(gistURL string) error {
	// Parse gist ID from URL
	gistID := extractGistID(gistURL)
	if gistID == "" {
		return fmt.Errorf("invalid gist URL: %s", gistURL)
	}

	// Fetch gist
	apiURL := fmt.Sprintf("https://api.github.com/gists/%s", gistID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	// Add token if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	fmt.Println("Fetching configuration...")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch gist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch gist (status %d)", resp.StatusCode)
	}

	var gist struct {
		Files map[string]struct {
			Content string `json:"content"`
		} `json:"files"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&gist); err != nil {
		return fmt.Errorf("failed to decode gist: %w", err)
	}

	// Find config file
	var configContent string
	for name, file := range gist.Files {
		if strings.HasSuffix(name, ".toml") || name == "config" {
			configContent = file.Content
			break
		}
	}

	if configContent == "" {
		return fmt.Errorf("no configuration file found in gist")
	}

	// Backup and apply
	homeDir := global.HomeDir()
	configPath := filepath.Join(homeDir, "config")

	// Backup existing config
	if _, err := os.Stat(configPath); err == nil {
		backupPath := configPath + ".backup"
		data, _ := os.ReadFile(configPath)
		_ = os.WriteFile(backupPath, data, global.ModeFile)
		fmt.Printf("Backed up existing config to: %s\n", backupPath)
	}

	// Write new config
	if err := os.WriteFile(configPath, []byte(configContent), global.ModeFile); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Handle base64 encoded image if present
	for name, file := range gist.Files {
		if strings.HasSuffix(name, ".base64.txt") {
			imageName := strings.TrimSuffix(name, ".base64.txt")
			if err := s.decodeAndSaveImage(file.Content, imageName); err != nil {
				fmt.Printf("Warning: could not decode image: %v\n", err)
			}
		}
	}

	successColor := color.New(color.FgGreen, color.Bold)
	successColor.Println("Configuration imported successfully!")
	fmt.Println()
	fmt.Println("Note: You may need to update the image path in the config file.")
	fmt.Println("Run 'chafa-welcome' to see your new configuration.")

	return nil
}

// decodeAndSaveImage decodes and saves a base64 encoded image
func (s *Share) decodeAndSaveImage(content, imageName string) error {
	// Remove header comments
	lines := strings.Split(content, "\n")
	var base64Data string
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
			base64Data += strings.TrimSpace(line)
		}
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// Save to images directory
	imagesDir := filepath.Join(global.HomeDir(), "images")
	if err := os.MkdirAll(imagesDir, global.ModeDir); err != nil {
		return err
	}

	imagePath := filepath.Join(imagesDir, imageName)
	if err := os.WriteFile(imagePath, data, global.ModeFile); err != nil {
		return err
	}

	fmt.Printf("Image saved to: %s\n", imagePath)
	return nil
}

// extractGistID extracts the gist ID from a URL
func extractGistID(url string) string {
	// Handle various URL formats:
	// https://gist.github.com/username/gist_id
	// https://gist.github.com/gist_id
	// gist_id (raw ID)

	url = strings.TrimSpace(url)

	// If it's just an ID (alphanumeric)
	if !strings.Contains(url, "/") && !strings.Contains(url, ".") {
		return url
	}

	// Parse URL
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}
