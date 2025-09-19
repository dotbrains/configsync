package util

import (
	"os"
)

// PathExists checks if a path exists on the filesystem
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}