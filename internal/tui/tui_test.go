package tui

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/WagnerMatos/semver/internal/config"
	"github.com/WagnerMatos/semver/internal/version"
	"github.com/charmbracelet/bubbles/textinput"
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
			name:       "confirm changes",
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")},
			initState:  stateConfirm,
			wantState:  stateTagConfirm,
			cursor:     0,
			wantCursor: 0,
		},
		{
			name:       "confirm tag",
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")},
			initState:  stateTagConfirm,
			wantState:  stateTagConfirm,
			cursor:     0,
			wantCursor: 0,
		},
		{
			name:       "cancel tag",
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")},
			initState:  stateTagConfirm,
			wantState:  stateTagConfirm,
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

func TestSaveChanges(t *testing.T) {
	tests := []struct {
		name      string
		bumpErr   error
		readErr   error
		updateErr error
		commitErr error
		tagErr    error
		createTag bool
		wantErr   bool
	}{
		{
			name:      "successful save without tag",
			createTag: false,
			wantErr:   false,
		},
		{
			name:      "successful save with tag",
			createTag: true,
			wantErr:   false,
		},
		{
			name:    "bump error",
			bumpErr: errTest,
			wantErr: true,
		},
		{
			name:    "read error",
			readErr: errTest,
			wantErr: true,
		},
		{
			name:      "update error",
			updateErr: errTest,
			wantErr:   true,
		},
		{
			name:      "commit error",
			commitErr: errTest,
			wantErr:   true,
		},
		{
			name:      "tag error",
			createTag: true,
			tagErr:    errTest,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				cfg:    &config.Config{},
				logger: slog.Default(),
				version: &mockVersionService{
					version: &version.Version{1, 0, 0},
					bumpErr: tt.bumpErr,
					readErr: tt.readErr,
				},
				git: &mockGitService{
					commitErr: tt.commitErr,
					tagErr:    tt.tagErr,
				},
				log: &mockChangelogService{
					updateErr: tt.updateErr,
				},
			}

			m := &model{
				ctx:        context.Background(),
				app:        app,
				commitType: version.Major,
				shortDesc:  textinput.New(),
				longDesc:   textinput.New(),
			}

			err := m.saveChanges(tt.createTag)
			if (err != nil) != tt.wantErr {
				t.Errorf("saveChanges() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateTag(t *testing.T) {
	tests := []struct {
		name    string
		readErr error
		tagErr  error
		wantErr bool
	}{
		{
			name:    "successful tag creation",
			wantErr: false,
		},
		{
			name:    "read error",
			readErr: errTest,
			wantErr: true,
		},
		{
			name:    "tag error",
			tagErr:  errTest,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				cfg:    &config.Config{},
				logger: slog.Default(),
				version: &mockVersionService{
					version: &version.Version{1, 0, 0},
					readErr: tt.readErr,
				},
				git: &mockGitService{
					tagErr: tt.tagErr,
				},
			}

			m := &model{
				ctx: context.Background(),
				app: app,
			}

			err := m.createTag()
			if (err != nil) != tt.wantErr {
				t.Errorf("createTag() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var errTest = errors.New("test error")

