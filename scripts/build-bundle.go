package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type bundleFile struct {
	Manifest   json.RawMessage   `json:"manifest"`
	Assets     []bundleAsset     `json:"assets,omitempty"`
	Migrations []bundleMigration `json:"migrations,omitempty"`
}

type bundleAsset struct {
	Path        string `json:"path"`
	Content     string `json:"content"`
	ContentType string `json:"contentType,omitempty"`
}

type bundleMigration struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func main() {
	source := flag.String("source", ".", "Extension source directory")
	out := flag.String("out", "", "Output bundle path")
	flag.Parse()

	if strings.TrimSpace(*out) == "" {
		exitf("missing --out")
	}

	bundle, err := buildBundle(filepath.Clean(*source))
	if err != nil {
		exitf("%v", err)
	}
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		exitf("encode bundle: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(filepath.Clean(*out)), 0o755); err != nil {
		exitf("create output directory: %v", err)
	}
	if err := os.WriteFile(filepath.Clean(*out), append(data, '\n'), 0o644); err != nil {
		exitf("write bundle: %v", err)
	}
}

func buildBundle(root string) (bundleFile, error) {
	manifestPath := filepath.Join(root, "manifest.json")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return bundleFile{}, fmt.Errorf("read manifest: %w", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return bundleFile{}, fmt.Errorf("decode manifest: %w", err)
	}

	assets, err := collectAssets(filepath.Join(root, "assets"))
	if err != nil {
		return bundleFile{}, err
	}
	migrations, err := collectMigrations(filepath.Join(root, "migrations"))
	if err != nil {
		return bundleFile{}, err
	}

	return bundleFile{
		Manifest:   manifestBytes,
		Assets:     assets,
		Migrations: migrations,
	}, nil
}

func collectAssets(root string) ([]bundleAsset, error) {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat assets: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("assets path is not a directory: %s", root)
	}

	assets := []bundleAsset{}
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		assets = append(assets, bundleAsset{
			Path:        filepath.ToSlash(relative),
			Content:     string(content),
			ContentType: detectContentType(path, content),
		})
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk assets: %w", err)
	}

	sort.Slice(assets, func(i, j int) bool {
		return assets[i].Path < assets[j].Path
	})
	return assets, nil
}

func collectMigrations(root string) ([]bundleMigration, error) {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat migrations: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("migrations path is not a directory: %s", root)
	}

	migrations := []bundleMigration{}
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		migrations = append(migrations, bundleMigration{
			Path:    filepath.ToSlash(relative),
			Content: string(content),
		})
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk migrations: %w", err)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Path < migrations[j].Path
	})
	return migrations, nil
}

func detectContentType(path string, content []byte) string {
	if ext := strings.ToLower(filepath.Ext(path)); ext != "" {
		if byExt := mime.TypeByExtension(ext); byExt != "" {
			return byExt
		}
	}
	return http.DetectContentType(content)
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
