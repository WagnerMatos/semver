package tui

import (
	"context"
	"log/slog"
	"testing"

	"github.com/WagnerMatos/semver/internal/config"
	"github.com/WagnerMatos/semver/internal/version"
	tea "github.com/charmbracelet/bubbletea"
)

type mockVersionService struct {
	version      *version.Version
	bumpErr      error
	readErr      error
	writeErr     error
	getLatestErr error
}

func (m *mockVersionService) GetLatestVersion() (*version.Version, error) {
	if m.getLatestErr != nil {
		return nil, m.getLatestErr
	}
	return m.version, nil
}

func (m *mockVersionService) Read() (*version.Version, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	return m.version, nil
}

func (m *mockVersionService) Write(v *version.Version) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.version = v
	return nil
}

func (m *mockVersionService) Bump(t version.Type) error {
	if m.bumpErr != nil {
		return m.bumpErr
	}
	return nil
}

type mockGitService struct {
	commitErr error
	tagErr    error
}

func (m *mockGitService) Commit(ctx context.Context, message string) error {
	return m.commitErr
}

func (m *mockGitService) Tag(ctx context.Context, ver *version.Version) error {
	return m.tagErr
}

type mockChangelogService struct {
	updateErr error
}

func (m *mockChangelogService) Update(v version.Version, t version.Type, shortDesc, longDesc string) error {
	return m.updateErr
}

func TestModel_Update(t *testing.T) {
	tests := []struct {
		name       string
		msg        tea.Msg
		initState  state
		wantState  state
		cursor     int
		wantCursor int
		setupModel func(m *model) // New setup function for model
	}{
		{
			name:       "move cursor up",
			msg:        tea.KeyMsg{Type: tea.KeyUp},
			initState:  stateCommitType,
			wantState:  stateCommitType,
			cursor:     1,
			wantCursor: 0,
		},
		{
			name:       "move cursor down",
			msg:        tea.KeyMsg{Type: tea.KeyDown},
			initState:  stateCommitType,
			wantState:  stateCommitType,
			cursor:     0,
			wantCursor: 1,
		},
		{
			name:       "select commit type",
			msg:        tea.KeyMsg{Type: tea.KeyEnter},
			initState:  stateCommitType,
			wantState:  stateShortDesc,
			cursor:     0,
			wantCursor: 0,
		},
		{
			name:       "enter short description",
			msg:        tea.KeyMsg{Type: tea.KeyEnter},
			initState:  stateShortDesc,
			wantState:  stateLongDesc,
			cursor:     0,
			wantCursor: 0,
			setupModel: func(m *model) {
				m.shortDesc.SetValue("A valid short description")
			},
		},
		{
			name:       "enter long description",
			msg:        tea.KeyMsg{Type: tea.KeyEnter},
			initState:  stateLongDesc,
			wantState:  stateTagConfirm,
			cursor:     0,
			wantCursor: 0,
		},
		{
			name:       "confirm tag creation",
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")},
			initState:  stateTagConfirm,
			wantState:  stateConfirm,
			cursor:     0,
			wantCursor: 0,
		},
		{
			name:       "confirm changes",
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")},
			initState:  stateConfirm,
			wantState:  stateConfirm,
			cursor:     0,
			wantCursor: 0,
		},
		{
			name:       "cancel changes",
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")},
			initState:  stateConfirm,
			wantState:  stateConfirm,
			cursor:     0,
			wantCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				cfg:     &config.Config{},
				logger:  slog.Default(),
				version: &mockVersionService{version: &version.Version{1, 0, 0}},
				git:     &mockGitService{},
				log:     &mockChangelogService{},
			}

			m := initialModel(context.Background(), app)
			m.state = tt.initState
			m.cursor = tt.cursor

			if tt.setupModel != nil {
				tt.setupModel(&m) // Apply any setup before updating
			}

			newModel, _ := m.Update(tt.msg)
			updatedModel := newModel.(model)

			if updatedModel.state != tt.wantState && !updatedModel.quitting {
				t.Errorf("state = %v, want %v", updatedModel.state, tt.wantState)
			}

			if updatedModel.cursor != tt.wantCursor {
				t.Errorf("cursor = %v, want %v", updatedModel.cursor, tt.wantCursor)
			}
		})
	}
}

