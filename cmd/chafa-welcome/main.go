package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/av1ppp/chafa-welcome/internal/chafa"
	"github.com/av1ppp/chafa-welcome/internal/config"
	"github.com/av1ppp/chafa-welcome/internal/gallery"
	"github.com/av1ppp/chafa-welcome/internal/global"
	"github.com/av1ppp/chafa-welcome/internal/share"
	"github.com/av1ppp/chafa-welcome/internal/sysinfo"
)

var (
	// Version information (can be set via ldflags)
	version = "dev"
)

func main() {
	if err := innerMain(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func innerMain() error {
	// Define flags
	showGallery := flag.Bool("gallery", false, "Browse community presets gallery")
	showCommunity := flag.Bool("community", false, "Browse community presets gallery (alias for --gallery)")
	shareConfig := flag.Bool("share", false, "Share current configuration to GitHub Gist")
	sharePublic := flag.Bool("public", false, "Make shared gist public (used with --share)")
	shareWithImage := flag.Bool("with-image", false, "Include image when sharing (used with --share)")
	importGist := flag.String("import", "", "Import configuration from a GitHub Gist URL")
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help information")

	flag.Parse()

	// Handle --help
	if *showHelp {
		printHelp()
		return nil
	}

	// Handle --version
	if *showVersion {
		fmt.Printf("chafa-welcome version %s\n", version)
		return nil
	}

	// Handle --gallery or --community
	if *showGallery || *showCommunity {
		g := gallery.New()
		return g.Run()
	}

	// Handle --share
	if *shareConfig {
		s := share.New()
		opts := share.ShareOptions{
			IncludeImage: *shareWithImage,
			Public:       *sharePublic,
			UseGhCli:     true, // Prefer gh CLI
		}
		return s.Run(opts)
	}

	// Handle --import
	if *importGist != "" {
		s := share.New()
		return s.Import(*importGist)
	}

	// Default: show system info with ASCII art
	return showWelcome()
}

func showWelcome() error {
	homeDir := global.HomeDir()
	confPath := filepath.Join(homeDir, "config")
	conf, err := config.ParseFile(confPath)
	if err != nil {
		return err
	}

	gap := strings.Repeat(" ", conf.Body.Gap)
	pictureMarginLeft := strings.Repeat(" ", conf.Offset.X)

	info, err := sysinfo.Collect(conf)
	if err != nil {
		return fmt.Errorf("failed to collect system info: %s", err.Error())
	}
	infoLines := strings.Split(info.String(), "\n")
	infoNumberLines := len(infoLines)

	chafaOutput, err := chafa.Execute(conf)
	if err != nil {
		return err
	}
	chafaLines := strings.Split(chafaOutput, "\n")
	chafaNumberLines := len(chafaLines) - 1
	chafaEmptyRow := strings.Repeat(" ", conf.Image.Size)

	maxLines := 0
	if infoNumberLines > chafaNumberLines {
		maxLines = infoNumberLines
	} else {
		maxLines = chafaNumberLines
	}

	resultBuilder := strings.Builder{}

	for i := 0; i < conf.Offset.Y; i++ {
		resultBuilder.WriteByte('\n')
	}

	for i := 0; i < maxLines; i++ {
		if i < chafaNumberLines {
			// with picture row
			if i < infoNumberLines {
				// with info row
				resultBuilder.WriteString(pictureMarginLeft + chafaLines[i] + gap + infoLines[i] + "\n")
			} else {
				// without info row
				resultBuilder.WriteString(pictureMarginLeft + chafaLines[i] + "\n")
			}
		} else {
			// without picture row
			resultBuilder.WriteString(pictureMarginLeft + chafaEmptyRow + gap + infoLines[i] + "\n")
		}
	}

	fmt.Println(resultBuilder.String())
	return nil
}

func printHelp() {
	fmt.Println("chafa-welcome - System information with ASCII art")
	fmt.Println()
	fmt.Println("Usage: chafa-welcome [OPTIONS]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --help            Show this help message")
	fmt.Println("  --version         Show version information")
	fmt.Println()
	fmt.Println("Community Features:")
	fmt.Println("  --gallery         Browse community presets gallery")
	fmt.Println("  --community       Alias for --gallery")
	fmt.Println()
	fmt.Println("Sharing:")
	fmt.Println("  --share           Share current configuration to GitHub Gist")
	fmt.Println("  --public          Make the shared gist public (default: private)")
	fmt.Println("  --with-image      Include the image when sharing")
	fmt.Println("  --import URL      Import configuration from a GitHub Gist URL")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  Config file: ~/.chafa-welcome/config")
	fmt.Println("  Gallery cache: ~/.chafa-welcome/gallery/")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  chafa-welcome                    Show system info with ASCII art")
	fmt.Println("  chafa-welcome --gallery          Browse community presets")
	fmt.Println("  chafa-welcome --share            Share your config to GitHub Gist")
	fmt.Println("  chafa-welcome --share --public   Share as a public gist")
	fmt.Println("  chafa-welcome --import URL       Import config from a gist")
	fmt.Println()
	fmt.Println("For more information, visit:")
	fmt.Println("  https://github.com/av1ppp/chafa-welcome")
}
