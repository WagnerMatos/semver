package version

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersion_Bump(t *testing.T) {
	tests := []struct {
		name        string
		version     Version
		versionType Type
		want        Version
		wantErr     bool
	}{
		{
			name:        "bump major version",
			version:     Version{1, 2, 3},
			versionType: Major,
			want:        Version{2, 0, 0},
			wantErr:     false,
		},
		{
			name:        "bump minor version",
			version:     Version{1, 2, 3},
			versionType: Minor,
			want:        Version{1, 3, 0},
			wantErr:     false,
		},
		{
			name:        "bump patch version",
			version:     Version{1, 2, 3},
			versionType: Patch,
			want:        Version{1, 2, 4},
			wantErr:     false,
		},
		{
			name:        "invalid version type",
			version:     Version{1, 2, 3},
			versionType: "invalid",
			want:        Version{1, 2, 3},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.version.Bump(tt.versionType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bump() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (tt.version != tt.want) {
				t.Errorf("Bump() = %v, want %v", tt.version, tt.want)
			}
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	tests := []struct {
		name     string
		version1 Version
		version2 Version
		want     int
	}{
		{
			name:     "v1 < v2 (major)",
			version1: Version{1, 0, 0},
			version2: Version{2, 0, 0},
			want:     -1,
		},
		{
			name:     "v1 > v2 (major)",
			version1: Version{2, 0, 0},
			version2: Version{1, 0, 0},
			want:     1,
		},
		{
			name:     "v1 < v2 (minor)",
			version1: Version{1, 1, 0},
			version2: Version{1, 2, 0},
			want:     -1,
		},
		{
			name:     "v1 > v2 (minor)",
			version1: Version{1, 2, 0},
			version2: Version{1, 1, 0},
			want:     1,
		},
		{
			name:     "v1 < v2 (patch)",
			version1: Version{1, 1, 1},
			version2: Version{1, 1, 2},
			want:     -1,
		},
		{
			name:     "v1 > v2 (patch)",
			version1: Version{1, 1, 2},
			version2: Version{1, 1, 1},
			want:     1,
		},
		{
			name:     "v1 = v2",
			version1: Version{1, 1, 1},
			version2: Version{1, 1, 1},
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version1.Compare(&tt.version2)
			if got != tt.want {
				t.Errorf("Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createFile(t *testing.T, path string, content string) {
	t.Helper()
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
}

func TestFileService_GetLatestVersion(t *testing.T) {
	dir := t.TempDir()
	versionFile := filepath.Join(dir, "VERSION.md")
	changelogFile := filepath.Join(dir, "CHANGELOG.md")

	tests := []struct {
		name             string
		versionContent   string
		changelogContent string
		setupFiles       func()
		want             *Version
		wantErr          bool
	}{
		{
			name: "version from VERSION.md",
			setupFiles: func() {
				createFile(t, versionFile, "1.2.3")
			},
			want:    &Version{1, 2, 3},
			wantErr: false,
		},
		{
			name: "version from CHANGELOG.md",
			setupFiles: func() {
				createFile(t, changelogFile, `## [2.0.0] - 2024-12-23
### Major
- desc
## [1.0.0] - 2024-12-23
### Major
- Initial commit`)
			},
			want:    &Version{2, 0, 0},
			wantErr: false,
		},
		{
			name:    "no files exist",
			want:    &Version{0, 1, 0},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up files from previous test
			os.Remove(versionFile)
			os.Remove(changelogFile)

			if tt.setupFiles != nil {
				tt.setupFiles()
			}

			fs := NewFileService(versionFile)
			got, err := fs.GetLatestVersion()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Compare(tt.want) != 0 {
				t.Errorf("GetLatestVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileService_Bump(t *testing.T) {
	dir := t.TempDir()
	versionFile := filepath.Join(dir, "VERSION.md")
	changelogFile := filepath.Join(dir, "CHANGELOG.md")

	tests := []struct {
		name       string
		setupFiles func()
		bumpType   Type
		want       string
		wantErr    bool
	}{
		{
			name: "bump major from changelog",
			setupFiles: func() {
				createFile(t, changelogFile, `## [2.0.0] - 2024-12-23
### Major
- desc`)
			},
			bumpType: Major,
			want:     "3.0.0",
			wantErr:  false,
		},
		{
			name: "bump minor from changelog",
			setupFiles: func() {
				createFile(t, changelogFile, `## [2.0.0] - 2024-12-23
### Major
- desc`)
			},
			bumpType: Minor,
			want:     "2.1.0",
			wantErr:  false,
		},
		{
			name: "bump patch from changelog",
			setupFiles: func() {
				createFile(t, changelogFile, `## [2.0.0] - 2024-12-23
### Major
- desc`)
			},
			bumpType: Patch,
			want:     "2.0.1",
			wantErr:  false,
		},
		{
			name: "bump from version file",
			setupFiles: func() {
				createFile(t, versionFile, "1.2.3")
			},
			bumpType: Minor,
			want:     "1.3.0",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up files from previous test
			os.Remove(versionFile)
			os.Remove(changelogFile)

			if tt.setupFiles != nil {
				tt.setupFiles()
			}

			fs := NewFileService(versionFile)
			err := fs.Bump(tt.bumpType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bump() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				content, err := os.ReadFile(versionFile)
				if err != nil {
					t.Fatalf("Failed to read version file: %v", err)
				}
				if got := string(content); got != tt.want {
					t.Errorf("Version file content = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

