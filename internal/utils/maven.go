package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// RunMavenBuild executes 'mvn clean install -DskipTests' in the given directory
func RunMavenBuild(projectDir string) error {
	cmd := exec.Command("mvn", "clean", "install", "-DskipTests", "-q")
	cmd.Dir = projectDir

	// Capture output for error reporting
	output, err := cmd.CombinedOutput()

	if err != nil {
		if len(output) > 0 {
			// Return last 20 lines of output
			lines := strings.Split(string(output), "\n")
			start := 0
			if len(lines) > 20 {
				start = len(lines) - 20
			}
			return fmt.Errorf("%w\n\nMaven output:\n%s", err, strings.Join(lines[start:], "\n"))
		}
		return err
	}

	return nil
}

// RunMavenCompile executes 'mvn clean compile -DskipTests' in the given directory
func RunMavenCompile(projectDir string) error {
	cmd := exec.Command("mvn", "clean", "compile", "-DskipTests", "-q")
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()

	if err != nil {
		if len(output) > 0 {
			lines := strings.Split(string(output), "\n")
			start := 0
			if len(lines) > 20 {
				start = len(lines) - 20
			}
			return fmt.Errorf("%w\n\nMaven output:\n%s", err, strings.Join(lines[start:], "\n"))
		}
		return err
	}

	return nil
}

// IsMavenAvailable checks if Maven is available on the system
func IsMavenAvailable() bool {
	cmd := exec.Command("mvn", "--version")
	err := cmd.Run()
	return err == nil
}
