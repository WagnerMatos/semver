// internal/version/version.go
package version

import (
	"errors"
	"fmt"
	"os"
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
}

type FileService struct {
	filepath string
	version  *Version
}

func NewFileService(filepath string) *FileService {
	return &FileService{
		filepath: filepath,
		version:  &Version{0, 1, 0},
	}
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
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

	var major, minor, patch int
	_, err = fmt.Sscanf(string(data), "%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidVersion, err)
	}

	s.version = &Version{major, minor, patch}
	return s.version, nil
}

func (s *FileService) Write(v *Version) error {
	if err := os.WriteFile(s.filepath, []byte(v.String()), 0644); err != nil {
		return fmt.Errorf("writing version file: %w", err)
	}
	return nil
}

func (s *FileService) Bump(t Type) error {
	if err := s.version.Bump(t); err != nil {
		return fmt.Errorf("bumping version: %w", err)
	}
	return s.Write(s.version)
}

