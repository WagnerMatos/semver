package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WagnerMatos/semver/internal/version"
)

func setupGitRepo(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	configCmds := [][]string{
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, args := range configCmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to configure git: %v", err)
		}
	}

	return tempDir
}

func getGitTags(t *testing.T, dir string) []string {
	t.Helper()
	cmd := exec.Command("git", "tag")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get git tags: %v", err)
	}
	if len(output) == 0 {
		return []string{}
	}
	return strings.Split(strings.TrimSpace(string(output)), "\n")
}

func TestGitService_Tag(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	dir := setupGitRepo(t)
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	s := New()
	if err := s.Commit(context.Background(), "initial commit"); err != nil {
		t.Fatalf("Failed to make initial commit: %v", err)
	}

	tests := []struct {
		name    string
		version *version.Version
		wantTag string
		wantErr bool
	}{
		{
			name:    "valid tag",
			version: &version.Version{Major: 1, Minor: 0, Patch: 0},
			wantTag: "v1.0.0",
			wantErr: false,
		},
		{
			name:    "another valid tag",
			version: &version.Version{Major: 2, Minor: 1, Patch: 3},
			wantTag: "v2.1.3",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Tag(context.Background(), tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				tags := getGitTags(t, dir)
				found := false
				for _, tag := range tags {
					if tag == tt.wantTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Tag() tag %s not found in repository", tt.wantTag)
				}
			}
		})
	}
}

func TestGitService_Commit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	dir := setupGitRepo(t)
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{
			name:    "valid commit",
			message: "test commit",
			wantErr: false,
		},
		{
			name:    "empty message",
			message: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			err := s.Commit(context.Background(), tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("Commit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

