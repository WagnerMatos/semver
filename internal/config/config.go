package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	VersionFile   string
	ChangelogFile string
}

func Load() (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return &Config{
		VersionFile:   filepath.Join(wd, "VERSION.md"),
		ChangelogFile: filepath.Join(wd, "CHANGELOG.md"),
	}, nil
}
