package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	// Setup
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	// Run tests
	code := m.Run()
	os.Exit(code)
}

func TestRunFunction(t *testing.T) {
	// Create a test directory
	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Setup logger with test handler
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create test files
	testCases := []struct {
		name            string
		setupFiles      func(t *testing.T)
		expectError     bool
		expectedFiles   []string
		unexpectedFiles []string
	}{
		{
			name:       "clean directory",
			setupFiles: func(t *testing.T) {},
			expectedFiles: []string{
				"VERSION.md",
				"CHANGELOG.md",
			},
		},
		{
			name: "existing version file",
			setupFiles: func(t *testing.T) {
				err := os.WriteFile(filepath.Join(dir, "VERSION.md"), []byte("1.0.0"), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			},
			expectedFiles: []string{
				"VERSION.md",
				"CHANGELOG.md",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean directory before each test
			files, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("Failed to read directory: %v", err)
			}
			for _, file := range files {
				os.Remove(filepath.Join(dir, file.Name()))
			}

			// Setup test files
			tc.setupFiles(t)

			// Run the function in test mode
			ctx := context.Background()
			err = run(ctx, logger, true) // Set testing to true

			// Check error
			if (err != nil) != tc.expectError {
				t.Errorf("run() error = %v, expectError %v", err, tc.expectError)
			}

			// Check expected files exist
			for _, file := range tc.expectedFiles {
				path := filepath.Join(dir, file)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("Expected file %s does not exist", file)
				}
			}

			// Check unexpected files don't exist
			for _, file := range tc.unexpectedFiles {
				path := filepath.Join(dir, file)
				if _, err := os.Stat(path); err == nil {
					t.Errorf("Unexpected file %s exists", file)
				}
			}
		})
	}
}

