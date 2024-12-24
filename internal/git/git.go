package git

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/WagnerMatos/semver/internal/version"
)

var (
	ErrCommitFailed = errors.New("commit failed")
	ErrAddFailed    = errors.New("add failed")
	ErrTagFailed    = errors.New("tag failed")
)

type Service interface {
	Commit(context.Context, string) error
	Tag(context.Context, *version.Version) error
}

type GitService struct{}

func New() *GitService {
	return &GitService{}
}

func (s *GitService) Commit(ctx context.Context, message string) error {
	if err := s.add(ctx); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "git", "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %v", ErrCommitFailed, err)
	}

	return nil
}

func (s *GitService) Tag(ctx context.Context, ver *version.Version) error {
	tagName := fmt.Sprintf("v%s", ver.String())
	cmd := exec.CommandContext(ctx, "git", "tag", tagName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %v", ErrTagFailed, err)
	}

	return nil
}

func (s *GitService) add(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "add", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %v", ErrAddFailed, err)
	}
	return nil
}

