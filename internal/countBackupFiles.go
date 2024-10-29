package internal

import (
	"os"
)

func CountBackupFiles(path string) ([]string, error) {
	var backups []string
	dir, err := os.ReadDir(path)
	if err != nil {
		return backups, err
	}
	for _, file := range dir {
		backups = append(backups, file.Name())
	}

	return backups, nil

}
