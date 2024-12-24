package tui

import (
	"context"
	"errors"
	"log/slog"
	"strings"
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
}

func (m *mockGitService) Commit(ctx context.Context, message string) error {
	return m.commitErr
}

type mockChangelogService struct {
	updateErr error
}

func (m *mockChangelogService) Update(v version.Version, t version.Type, shortDesc, longDesc string) error {
	return m.updateErr
}

var errTest = errors.New("test error")

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
			name:       "quit",
			msg:        tea.KeyMsg{Type: tea.KeyCtrlC},
			initState:  stateCommitType,
			wantState:  stateCommitType,
			cursor:     0,
			wantCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{
				cfg:     &config.Config{},
				logger:  slog.Default(),
				version: &mockVersionService{},
				git:     &mockGitService{},
				log:     &mockChangelogService{},
			}

			m := initialModel(context.Background(), app)
			m.state = tt.initState
			m.cursor = tt.cursor

			newModel, _ := m.Update(tt.msg)
			updatedModel := newModel.(model)

			if updatedModel.state != tt.wantState {
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
		wantErr   bool
	}{
		{
			name:    "successful save",
			wantErr: false,
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
				},
				log: &mockChangelogService{
					updateErr: tt.updateErr,
				},
			}

			shortDesc := textinput.New()
			longDesc := textinput.New()

			m := &model{
				ctx:        context.Background(),
				app:        app,
				commitType: version.Major,
				shortDesc:  shortDesc,
				longDesc:   longDesc,
			}

			err := m.saveChanges()
			if (err != nil) != tt.wantErr {
				t.Errorf("saveChanges() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestView(t *testing.T) {
	tests := []struct {
		name     string
		state    state
		quitting bool
		err      error
		contains []string
	}{
		{
			name:     "commit type view",
			state:    stateCommitType,
			contains: []string{"Select commit type", "major", "minor", "patch"},
		},
		{
			name:     "short description view",
			state:    stateShortDesc,
			contains: []string{"Short description"},
		},
		{
			name:     "long description view",
			state:    stateLongDesc,
			contains: []string{"Long description"},
		},
		{
			name:     "confirm view",
			state:    stateConfirm,
			contains: []string{"confirm"},
		},
		{
			name:     "error view",
			quitting: true,
			err:      errTest,
			contains: []string{"Error: test error"},
		},
		{
			name:     "success view",
			quitting: true,
			contains: []string{"Changes saved successfully"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortDesc := textinput.New()
			longDesc := textinput.New()

			m := model{
				state:     tt.state,
				quitting:  tt.quitting,
				err:       tt.err,
				shortDesc: shortDesc,
				longDesc:  longDesc,
			}

			view := m.View()
			for _, s := range tt.contains {
				if !strings.Contains(strings.ToLower(view), strings.ToLower(s)) {
					t.Errorf("View() missing %q", s)
				}
			}
		})
	}
}
