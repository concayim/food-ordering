package main

import (
	"os"
	"path/filepath"
)

// dataDir 数据目录（SQLite、uploads），可通过环境变量 DATA_DIR 指定
func dataDir() string {
	if d := os.Getenv("DATA_DIR"); d != "" {
		return d
	}
	return "."
}

func dbPath() string {
	return filepath.Join(dataDir(), "food.db")
}

func uploadDirPath() string {
	return filepath.Join(dataDir(), "uploads")
}
