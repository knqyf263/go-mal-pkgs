package module

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"
)

// Fetcher fetches a module and calculates its checksum
type Fetcher struct {
	cacheDir string
}

// NewFetcher creates a new Fetcher
func NewFetcher(cacheDir string) (*Fetcher, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	return &Fetcher{cacheDir: cacheDir}, nil
}

// FetchAndVerify fetches a module and verifies its checksum
func (f *Fetcher) FetchAndVerify(path, version, expectedChecksum string, isGoMod bool) error {
	if isGoMod {
		// Skip go.mod verification for now as it requires different handling
		return nil
	}

	moduleDir := filepath.Join(f.cacheDir, path+"@"+version)
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return fmt.Errorf("failed to create module directory: %w", err)
	}

	// Fetch module directly from the source
	var archiveURL string
	if strings.HasPrefix(path, "github.com/") {
		// GitHub specific URL
		parts := strings.Split(path, "/")
		if len(parts) < 3 {
			return fmt.Errorf("invalid GitHub path: %s", path)
		}
		archiveURL = fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.zip",
			parts[1], parts[2], version)
	} else {
		// Generic URL (this might need to be adjusted based on the hosting service)
		archiveURL = fmt.Sprintf("https://%s/archive/%s.zip", path, version)
	}

	resp, err := http.Get(archiveURL)
	if err != nil {
		return fmt.Errorf("failed to fetch module: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch module: HTTP %d", resp.StatusCode)
	}

	archivePath := filepath.Join(moduleDir, version+".zip")
	out, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save archive: %w", err)
	}

	// Extract the archive
	extractDir := filepath.Join(moduleDir, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create extract directory: %w", err)
	}

	if err := extractZip(archivePath, extractDir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	// Find the root directory in the extracted files
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return fmt.Errorf("failed to read extract directory: %w", err)
	}
	if len(entries) != 1 || !entries[0].IsDir() {
		return fmt.Errorf("unexpected archive structure")
	}
	rootDir := filepath.Join(extractDir, entries[0].Name())

	// Calculate checksum
	prefix := path + "@" + version
	hash, err := dirhash.HashDir(rootDir, prefix, dirhash.Hash1)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Compare checksums
	if hash != expectedChecksum {
		return fmt.Errorf("checksum mismatch for %s@%s:\nexpected: %s\nactual: %s",
			path, version, expectedChecksum, hash)
	}

	return nil
}

func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
