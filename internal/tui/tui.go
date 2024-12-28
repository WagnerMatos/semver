package tui

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/WagnerMatos/semver/internal/changelog"
	"github.com/WagnerMatos/semver/internal/config"
	"github.com/WagnerMatos/semver/internal/git"
	"github.com/WagnerMatos/semver/internal/version"
)

type App struct {
	cfg     *config.Config
	logger  *slog.Logger
	version version.Service
	git     git.Service
	log     changelog.Service
	testing bool
}

func New(cfg *config.Config, logger *slog.Logger) *App {
	return &App{
		cfg:     cfg,
		logger:  logger,
		version: version.NewFileService(cfg.VersionFile),
		git:     git.New(),
		log:     changelog.New(cfg.ChangelogFile),
	}
}

func NewTest(cfg *config.Config, logger *slog.Logger) *App {
	app := New(cfg, logger)
	app.testing = true
	return app
}

func (a *App) Run(ctx context.Context) error {
	if a.testing {
		ver := &version.Version{0, 1, 0}
		if err := a.version.Write(ver); err != nil {
			return fmt.Errorf("writing initial version: %w", err)
		}

		f, err := os.OpenFile(a.cfg.ChangelogFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("creating changelog: %w", err)
		}
		f.Close()

		return nil
	}

	m := initialModel(ctx, a)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	return nil
}

type model struct {
	ctx        context.Context
	app        *App
	state      state
	cursor     int
	commitType version.Type
	shortDesc  textinput.Model
	longDesc   textinput.Model
	err        error
	quitting   bool
}

type state int

const (
	stateCommitType state = iota
	stateShortDesc
	stateLongDesc
	stateTagConfirm
	stateConfirm
)

var (
	commitTypes = []version.Type{version.Major, version.Minor, version.Patch}
	style       = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

func initialModel(ctx context.Context, app *App) model {
	shortDesc := textinput.New()
	shortDesc.Placeholder = "Enter short description"
	shortDesc.Focus()

	longDesc := textinput.New()
	longDesc.Placeholder = "Enter long description (optional)"

	return model{
		ctx:       ctx,
		app:       app,
		state:     stateCommitType,
		shortDesc: shortDesc,
		longDesc:  longDesc,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.state == stateCommitType {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(commitTypes) - 1
				}
			}

		case "down", "j":
			if m.state == stateCommitType {
				m.cursor++
				if m.cursor >= len(commitTypes) {
					m.cursor = 0
				}
			}

		case "y", "Y":
			switch m.state {
			case stateConfirm:
				if err := m.saveChanges(false); err != nil {
					m.err = err
					m.app.logger.Error("failed to save changes", "error", err)
					m.quitting = true
					return m, tea.Quit
				}
				m.quitting = true
				return m, tea.Quit
			case stateTagConfirm:
				if err := m.createTag(); err != nil {
					m.err = err
					m.app.logger.Error("failed to create tag", "error", err)
				}
				m.state = stateConfirm
			}

		case "n", "N":
			if m.state == stateConfirm || m.state == stateTagConfirm {
				m.quitting = true
				return m, tea.Quit
			}

		case "enter":
			switch m.state {
			case stateCommitType:
				m.commitType = commitTypes[m.cursor]
				m.state = stateShortDesc
			case stateShortDesc:
				if m.shortDesc.Value() != "" {
					m.state = stateLongDesc
				}
			case stateLongDesc:
				if err := m.app.version.Bump(m.commitType); err != nil {
					m.err = err
					m.app.logger.Error("failed to bump version", "error", err)
					m.quitting = true
					return m, tea.Quit
				}
				m.state = stateTagConfirm
			}
		}
	}

	if m.state == stateShortDesc {
		m.shortDesc, cmd = m.shortDesc.Update(msg)
	} else if m.state == stateLongDesc {
		m.longDesc, cmd = m.longDesc.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		if m.err != nil {
			return fmt.Sprintf("Error: %v\n", m.err)
		}
		return "Changes saved successfully!\n"
	}

	var s string
	switch m.state {
	case stateCommitType:
		s = "Select commit type (↑/↓ to move, enter to select):\n\n"
		for i, t := range commitTypes {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, t)
		}

	case stateShortDesc:
		s = "Short description:\n"
		s += m.shortDesc.View()

	case stateLongDesc:
		s = "Long description (optional):\n"
		s += m.longDesc.View()

	case stateTagConfirm:
		ver, _ := m.app.version.Read()
		s = fmt.Sprintf("\nCreate git tag v%s? (y/n)", ver.String())

	case stateConfirm:
		s = fmt.Sprintf("\nCommit Type: %s\nShort Description: %s\nLong Description: %s\n",
			m.commitType, m.shortDesc.Value(), m.longDesc.Value())
		s += "\nPress 'y' to confirm or 'n' to cancel"
	}

	return s
}

func (m *model) saveChanges(createTag bool) error {
	if err := m.app.version.Bump(m.commitType); err != nil {
		return fmt.Errorf("bumping version: %w", err)
	}

	ver, err := m.app.version.Read()
	if err != nil {
		return fmt.Errorf("reading version: %w", err)
	}

	if err := m.app.log.Update(*ver, m.commitType, m.shortDesc.Value(), m.longDesc.Value()); err != nil {
		return fmt.Errorf("updating changelog: %w", err)
	}

	if err := m.app.git.Commit(m.ctx, m.shortDesc.Value()); err != nil {
		return fmt.Errorf("committing changes: %w", err)
	}

	if createTag {
		if err := m.createTag(); err != nil {
			return fmt.Errorf("creating tag: %w", err)
		}
	}

	return nil
}

func (m *model) createTag() error {
	ver, err := m.app.version.Read()
	if err != nil {
		return fmt.Errorf("reading version: %w", err)
	}

	if err := m.app.git.Tag(m.ctx, ver); err != nil {
		return fmt.Errorf("creating tag: %w", err)
	}

	return nil
}

