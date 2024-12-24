package version

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	ErrInvalidVersion = errors.New("invalid version format")
	ErrInvalidType    = errors.New("invalid version type")
)

type Type string

const (
	Major Type = "major"
	Minor Type = "minor"
	Patch Type = "patch"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

type Service interface {
	Read() (*Version, error)
	Write(*Version) error
	Bump(Type) error
	GetLatestVersion() (*Version, error)
}

type FileService struct {
	filepath string
	version  *Version
}

func NewFileService(filepath string) *FileService {
	return &FileService{
		filepath: filepath,
		version:  &Version{0, 1, 0}, // Default version
	}
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func ParseVersion(s string) (*Version, error) {
	var major, minor, patch int
	_, err := fmt.Sscanf(s, "%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidVersion, err)
	}
	return &Version{major, minor, patch}, nil
}

// Compare returns:
//
//	-1 if v < other
//	 0 if v == other
//	 1 if v > other
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

func (v *Version) Bump(t Type) error {
	switch t {
	case Major:
		v.Major++
		v.Minor = 0
		v.Patch = 0
	case Minor:
		v.Minor++
		v.Patch = 0
	case Patch:
		v.Patch++
	default:
		return fmt.Errorf("%w: %s", ErrInvalidType, t)
	}
	return nil
}

func (s *FileService) Read() (*Version, error) {
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s.version, nil
		}
		return nil, fmt.Errorf("reading version file: %w", err)
	}

	ver, err := ParseVersion(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, err
	}

	s.version = ver
	return ver, nil
}

func (s *FileService) GetLatestVersion() (*Version, error) {
	// Try reading from VERSION.md first
	data, err := os.ReadFile(s.filepath)
	if err == nil {
		if ver, err := ParseVersion(strings.TrimSpace(string(data))); err == nil {
			s.version = ver
			return ver, nil
		}
	}

	// If VERSION.md doesn't exist or is invalid, try to parse from CHANGELOG.md
	changelogPath := filepath.Join(filepath.Dir(s.filepath), "CHANGELOG.md")
	data, err = os.ReadFile(changelogPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.version = &Version{0, 1, 0}
			return s.version, nil
		}
		return nil, fmt.Errorf("reading changelog: %w", err)
	}

	versions := make([]*Version, 0)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## [") && strings.Contains(line, "]") {
			verStr := strings.TrimPrefix(line, "## [")
			verStr = strings.Split(verStr, "]")[0]
			if ver, err := ParseVersion(strings.TrimSpace(verStr)); err == nil {
				versions = append(versions, ver)
			}
		}
	}

	if len(versions) == 0 {
		s.version = &Version{0, 1, 0}
		return s.version, nil
	}

	// Sort versions in descending order
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Compare(versions[j]) > 0
	})

	s.version = versions[0]
	return s.version, nil
}

func (s *FileService) Write(v *Version) error {
	if err := os.WriteFile(s.filepath, []byte(v.String()), 0644); err != nil {
		return fmt.Errorf("writing version file: %w", err)
	}
	return nil
}

func (s *FileService) Bump(t Type) error {
	latest, err := s.GetLatestVersion()
	if err != nil {
		return fmt.Errorf("getting latest version: %w", err)
	}

	s.version = latest
	if err := s.version.Bump(t); err != nil {
		return fmt.Errorf("bumping version: %w", err)
	}

	return s.Write(s.version)
}

