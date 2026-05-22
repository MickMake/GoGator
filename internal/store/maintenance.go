package store

import (
	"fmt"
	"path/filepath"
	"strings"
)

func Backup(sourcePath, destPath string) error {
	if strings.TrimSpace(destPath) == "" {
		return fmt.Errorf("missing backup path")
	}
	exists, err := Exists(sourcePath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("database not found: %s", sourcePath)
	}

	sourceAbs, err := filepath.Abs(sourcePath)
	if err != nil {
		return err
	}
	destAbs, err := filepath.Abs(destPath)
	if err != nil {
		return err
	}
	if sourceAbs == destAbs {
		return fmt.Errorf("backup path must differ from database path: %s", destPath)
	}

	destExists, err := Exists(destPath)
	if err != nil {
		return err
	}
	if destExists {
		return fmt.Errorf("backup already exists: %s", destPath)
	}

	db, err := Open(sourcePath)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec("VACUUM INTO " + quoteSQLiteString(destPath)); err != nil {
		return err
	}
	return nil
}

func Vacuum(path string) error {
	exists, err := Exists(path)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("database not found: %s", path)
	}

	db, err := Open(path)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("VACUUM")
	return err
}

func quoteSQLiteString(v string) string {
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}
