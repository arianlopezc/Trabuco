package utils

import (
	"os/exec"
	"strings"
)

// DockerStatus represents the status of Docker on the system
type DockerStatus struct {
	Installed bool
	Running   bool
	Version   string
	Error     string
}

// CheckDocker verifies that Docker is installed and running
func CheckDocker() DockerStatus {
	status := DockerStatus{}

	// Check if docker command exists
	_, err := exec.LookPath("docker")
	if err != nil {
		status.Error = "Docker is not installed or not in PATH"
		return status
	}
	status.Installed = true

	// Check if Docker daemon is running by executing 'docker info'
	cmd := exec.Command("docker", "info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Docker is installed but daemon is not running
		outputStr := string(output)
		if strings.Contains(outputStr, "Cannot connect") ||
			strings.Contains(outputStr, "Is the docker daemon running") ||
			strings.Contains(outputStr, "permission denied") {
			status.Error = "Docker daemon is not running. Please start Docker Desktop or the Docker service."
		} else {
			status.Error = "Docker daemon is not accessible: " + err.Error()
		}
		return status
	}
	status.Running = true

	// Get Docker version
	versionCmd := exec.Command("docker", "--version")
	versionOutput, err := versionCmd.Output()
	if err == nil {
		status.Version = strings.TrimSpace(string(versionOutput))
	}

	return status
}

// IsDockerReady returns true if Docker is installed and running
func IsDockerReady() bool {
	status := CheckDocker()
	return status.Installed && status.Running
}
