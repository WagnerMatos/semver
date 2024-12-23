// main.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

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

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter commit type (major/minor/patch): ")
	commitType, _ := reader.ReadString('\n')
	commitType = strings.TrimSpace(strings.ToLower(commitType))

	if commitType != "major" && commitType != "minor" && commitType != "patch" {
		fmt.Println("Invalid commit type. Must be major, minor, or patch")
		return
	}

	fmt.Print("Enter short description: ")
	shortDesc, _ := reader.ReadString('\n')
	shortDesc = strings.TrimSpace(shortDesc)

	fmt.Print("Enter long description (optional, press Enter to skip): ")
	longDesc, _ := reader.ReadString('\n')
	longDesc = strings.TrimSpace(longDesc)

	version, err := readVersion()
	if err != nil {
		fmt.Printf("Error reading version: %v\n", err)
		return
	}

	version.Bump(commitType)

	if err := writeVersion(version); err != nil {
		fmt.Printf("Error writing version: %v\n", err)
		return
	}

	if err := updateChangelog(commitType, shortDesc, longDesc, version); err != nil {
		fmt.Printf("Error updating changelog: %v\n", err)
		return
	}

	if err := gitCommit(shortDesc); err != nil {
		fmt.Printf("Error committing changes: %v\n", err)
		return
	}

	fmt.Printf("Successfully bumped version to %s and committed changes\n", version)
}
