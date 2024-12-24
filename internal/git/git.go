package git

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

var (
	ErrCommitFailed = errors.New("commit failed")
	ErrAddFailed    = errors.New("add failed")
)

type Service interface {
	Commit(context.Context, string) error
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

func (s *GitService) add(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "add", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %v", ErrAddFailed, err)
	}
	return nil
}
