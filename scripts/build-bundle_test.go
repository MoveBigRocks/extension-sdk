package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectMigrationsIgnoresNonSQLFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "000001_init.up.sql"), []byte("create table demo ();"), 0o600); err != nil {
		t.Fatalf("write sql migration: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "embed.go"), []byte("package migrations"), 0o600); err != nil {
		t.Fatalf("write helper file: %v", err)
	}

	migrations, err := collectMigrations(root)
	if err != nil {
		t.Fatalf("collect migrations: %v", err)
	}
	if len(migrations) != 1 {
		t.Fatalf("expected 1 bundled migration, got %d", len(migrations))
	}
	if migrations[0].Path != "000001_init.up.sql" {
		t.Fatalf("unexpected migration path %q", migrations[0].Path)
	}
}
