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

func TestFileService_ReadWrite(t *testing.T) {
	tempDir := t.TempDir()
	versionFile := filepath.Join(tempDir, "VERSION.md")

	tests := []struct {
		name     string
		content  string
		want     *Version
		wantErr  bool
		writeErr bool
	}{
		{
			name:    "read valid version",
			content: "1.2.3",
			want:    &Version{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "read invalid version",
			content: "invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "read empty file",
			content: "",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.content != "" {
				err := os.WriteFile(versionFile, []byte(tt.content), 0644)
				if err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			}

			fs := NewFileService(versionFile)
			got, err := fs.Read()
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() != tt.want.String() {
				t.Errorf("Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name    string
		version Version
		want    string
	}{
		{
			name:    "format version",
			version: Version{1, 2, 3},
			want:    "1.2.3",
		},
		{
			name:    "format version with zeros",
			version: Version{0, 0, 0},
			want:    "0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
