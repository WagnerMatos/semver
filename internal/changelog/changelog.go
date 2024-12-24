package changelog

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/WagnerMatos/semver/internal/version"
)

type Service interface {
	Update(version.Version, version.Type, string, string) error
}

type FileService struct {
	filepath string
}

func New(filepath string) *FileService {
	return &FileService{filepath: filepath}
}

func (s *FileService) Update(v version.Version, t version.Type, shortDesc, longDesc string) error {
	f, err := os.OpenFile(s.filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening changelog: %w", err)
	}
	defer f.Close()

	entry := fmt.Sprintf("\n## [%s] - %s\n", v.String(), time.Now().Format("2006-01-02"))
	entry += fmt.Sprintf("### %s\n", strings.Title(string(t)))
	entry += fmt.Sprintf("- %s\n", shortDesc)
	if longDesc != "" {
		entry += fmt.Sprintf("  %s\n", longDesc)
	}

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("writing changelog: %w", err)
	}

	return nil
}
