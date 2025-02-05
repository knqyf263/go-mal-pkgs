package sumdb

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Module represents a Go module with its checksum
type Module struct {
	Path     string
	Version  string
	Checksum string
	IsGoMod  bool // true if this is a go.mod file checksum
}

// Parse parses go.sum file and returns a slice of Module
func Parse(filename string) ([]Module, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open go.sum: %w", err)
	}
	defer file.Close()

	var modules []Module
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue
		}

		// Format: module-path version checksum
		// e.g. golang.org/x/text v0.3.0 h1:g61tztE5qeGQ89tm6NTjjM9VPIm088od1l6aSorWRWg=
		// or: golang.org/x/text v0.3.0/go.mod h1:g61tztE5qeGQ89tm6NTjjM9VPIm088od1l6aSorWRWg=
		path := parts[0]
		version := parts[1]
		isGoMod := false

		if strings.HasSuffix(version, "/go.mod") {
			isGoMod = true
			version = strings.TrimSuffix(version, "/go.mod")
		}

		modules = append(modules, Module{
			Path:     path,
			Version:  version,
			Checksum: parts[2],
			IsGoMod:  isGoMod,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan go.sum: %w", err)
	}

	return modules, nil
}
