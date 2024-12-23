package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	state      int
	cursor     int
	commitType string
	shortDesc  textinput.Model
	longDesc   textinput.Model
	err        error
	quitting   bool
}

const (
	stateCommitType = iota
	stateShortDesc
	stateLongDesc
	stateConfirm
)

var (
	commitTypes = []string{"major", "minor", "patch"}
	style       = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

func initialModel() model {
	shortDesc := textinput.New()
	shortDesc.Placeholder = "Enter short description"
	shortDesc.Focus()

	longDesc := textinput.New()
	longDesc.Placeholder = "Enter long description (optional)"

	return model{
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
			if m.state == stateConfirm {
				err := saveChanges(m.commitType, m.shortDesc.Value(), m.longDesc.Value())
				if err != nil {
					m.err = err
				}
				m.quitting = true
				return m, tea.Quit
			}
		case "n", "N":
			if m.state == stateConfirm {
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
				m.state = stateConfirm
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
	case stateConfirm:
		s = fmt.Sprintf("\nCommit Type: %s\nShort Description: %s\nLong Description: %s\n",
			m.commitType, m.shortDesc.Value(), m.longDesc.Value())
		s += "\nPress 'y' to confirm or 'n' to cancel"
	}

	return s
}

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *Version) Bump(commitType string) {
	switch commitType {
	case "major":
		v.Major++
		v.Minor = 0
		v.Patch = 0
	case "minor":
		v.Minor++
		v.Patch = 0
	case "patch":
		v.Patch++
	}
}

func readVersion() (*Version, error) {
	data, err := os.ReadFile("VERSION.md")
	if err != nil {
		if os.IsNotExist(err) {
			return &Version{0, 1, 0}, nil
		}
		return nil, err
	}

	var major, minor, patch int
	_, err = fmt.Sscanf(string(data), "%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return nil, err
	}

	return &Version{major, minor, patch}, nil
}

func writeVersion(v *Version) error {
	return os.WriteFile("VERSION.md", []byte(v.String()), 0644)
}

func updateChangelog(commitType, shortDesc, longDesc string, v *Version) error {
	f, err := os.OpenFile("CHANGELOG.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	entry := fmt.Sprintf("\n## [%s] - %s\n", v.String(), time.Now().Format("2006-01-02"))
	entry += fmt.Sprintf("### %s\n", strings.Title(commitType))
	entry += fmt.Sprintf("- %s\n", shortDesc)
	if longDesc != "" {
		entry += fmt.Sprintf("  %s\n", longDesc)
	}

	_, err = f.WriteString(entry)
	return err
}

func gitCommit(shortDesc string) error {
	cmd := exec.Command("git", "add", ".")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "commit", "-m", shortDesc)
	return cmd.Run()
}

func saveChanges(commitType, shortDesc, longDesc string) error {
	version, err := readVersion()
	if err != nil {
		return fmt.Errorf("error reading version: %v", err)
	}

	version.Bump(commitType)

	if err := writeVersion(version); err != nil {
		return fmt.Errorf("error writing version: %v", err)
	}

	if err := updateChangelog(commitType, shortDesc, longDesc, version); err != nil {
		return fmt.Errorf("error updating changelog: %v", err)
	}

	if err := gitCommit(shortDesc); err != nil {
		return fmt.Errorf("error committing changes: %v", err)
	}

	return nil
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

