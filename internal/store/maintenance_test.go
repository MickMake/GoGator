package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupCreatesReadableSQLiteCopy(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "gogator.sqlite")
	dest := filepath.Join(tmp, "backup.sqlite")

	if err := Init(source); err != nil {
		t.Fatalf("init source: %v", err)
	}
	if err := Backup(source, dest); err != nil {
		t.Fatalf("backup: %v", err)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("stat backup: %v", err)
	}
	if _, _, err := Status(dest); err != nil {
		t.Fatalf("status backup: %v", err)
	}
}

func TestBackupRefusesExistingDestination(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "gogator.sqlite")
	dest := filepath.Join(tmp, "backup.sqlite")

	if err := Init(source); err != nil {
		t.Fatalf("init source: %v", err)
	}
	if err := os.WriteFile(dest, []byte("already here"), 0o644); err != nil {
		t.Fatalf("write existing dest: %v", err)
	}
	if err := Backup(source, dest); err == nil {
		t.Fatalf("expected backup to refuse existing destination")
	}
}

func TestVacuumExistingDatabase(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "gogator.sqlite")

	if err := Init(path); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := Vacuum(path); err != nil {
		t.Fatalf("vacuum: %v", err)
	}
}
