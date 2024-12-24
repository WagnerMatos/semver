package config

import (
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !filepath.IsAbs(cfg.VersionFile) {
		t.Errorf("VersionFile is not absolute path: %v", cfg.VersionFile)
	}

	if !filepath.IsAbs(cfg.ChangelogFile) {
		t.Errorf("ChangelogFile is not absolute path: %v", cfg.ChangelogFile)
	}

	if filepath.Base(cfg.VersionFile) != "VERSION.md" {
		t.Errorf("VersionFile has wrong name: %v", cfg.VersionFile)
	}

	if filepath.Base(cfg.ChangelogFile) != "CHANGELOG.md" {
		t.Errorf("ChangelogFile has wrong name: %v", cfg.ChangelogFile)
	}
}

