package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/knqyf263/go-mal-pkgs/pkg/module"
	"github.com/knqyf263/go-mal-pkgs/pkg/sumdb"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("Please specify a project directory")
	}

	projectDir := flag.Arg(0)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		log.Fatalf("Project directory does not exist: %s", projectDir)
	}

	sumFile := filepath.Join(projectDir, "go.sum")
	if _, err := os.Stat(sumFile); os.IsNotExist(err) {
		log.Fatalf("go.sum file not found in: %s", projectDir)
	}

	if err := run(projectDir); err != nil {
		log.Fatal(err)
	}
}

func run(projectDir string) error {
	// Parse go.sum file
	modules, err := sumdb.Parse(filepath.Join(projectDir, "go.sum"))
	if err != nil {
		return fmt.Errorf("failed to parse go.sum: %w", err)
	}

	// Create a temporary directory for module cache
	tempDir, err := os.MkdirTemp("", "go-mal-pkgs-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create module fetcher
	fetcher, err := module.NewFetcher(tempDir)
	if err != nil {
		return fmt.Errorf("failed to create module fetcher: %w", err)
	}

	// Check each module
	for _, mod := range modules {
		if mod.IsGoMod {
			continue // Skip go.mod files for now
		}
		fmt.Printf("Checking %s@%s...\n", mod.Path, mod.Version)
		if err := fetcher.FetchAndVerify(mod.Path, mod.Version, mod.Checksum, mod.IsGoMod); err != nil {
			fmt.Printf("⚠️  WARNING: %v\n", err)
		} else {
			fmt.Printf("✅ Verified %s@%s\n", mod.Path, mod.Version)
		}
	}

	return nil
}
