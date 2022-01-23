package watcher

import (
	"os"
	"path/filepath"
)

func create(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0o770); err != nil {
		return nil, err
	}
	return os.Create(p)
}
