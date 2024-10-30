package internal

import (
	"os"
	"path/filepath"
)

func CountBackupFiles(path string) ([]string, error) {
	var backups []string
	dir, err := os.ReadDir(path)
	if err != nil {
		return backups, err
	}
	for _, file := range dir {
		if filepath.Ext(file.Name()) == ".bak" {
			backups = append(backups, file.Name())
		}

	}

	return backups, nil

}
