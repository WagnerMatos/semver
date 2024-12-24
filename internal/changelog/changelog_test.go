package changelog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WagnerMatos/semver/internal/version"
)

func TestFileService_Update(t *testing.T) {
	tempDir := t.TempDir()
	changelogFile := filepath.Join(tempDir, "CHANGELOG.md")

	tests := []struct {
		name      string
		version   version.Version
		vType     version.Type
		shortDesc string
		longDesc  string
		wantErr   bool
		check     func(t *testing.T, content string)
	}{
		{
			name:      "valid update with short description",
			version:   version.Version{Major: 1, Minor: 2, Patch: 3},
			vType:     version.Major,
			shortDesc: "test commit",
			longDesc:  "",
			wantErr:   false,
			check: func(t *testing.T, content string) {
				expected := []string{
					"[1.2.3]",
					"Major",
					"test commit",
				}
				for _, exp := range expected {
					if !strings.Contains(content, exp) {
						t.Errorf("Update() content missing %q", exp)
					}
				}
			},
		},
		{
			name:      "valid update with long description",
			version:   version.Version{Major: 1, Minor: 2, Patch: 3},
			vType:     version.Minor,
			shortDesc: "test commit",
			longDesc:  "long description",
			wantErr:   false,
			check: func(t *testing.T, content string) {
				expected := []string{
					"[1.2.3]",
					"Minor",
					"test commit",
					"long description",
				}
				for _, exp := range expected {
					if !strings.Contains(content, exp) {
						t.Errorf("Update() content missing %q", exp)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(changelogFile)
			err := s.Update(tt.version, tt.vType, tt.shortDesc, tt.longDesc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				content, err := os.ReadFile(changelogFile)
				if err != nil {
					t.Fatalf("Failed to read changelog file: %v", err)
				}
				tt.check(t, string(content))
			}
		})
	}
}
