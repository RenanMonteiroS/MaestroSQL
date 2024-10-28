package internal

import (
	"fmt"
	"os"
)

func CountBackupFiles(path string) (int, error) {
	dir, err := os.ReadDir("C:/_Backup")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(dir)
	return 0, nil
}
