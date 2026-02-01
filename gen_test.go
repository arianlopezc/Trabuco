//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/trabuco/trabuco/internal/config"
	"github.com/trabuco/trabuco/internal/generator"
)

func main() {
	cfg := &config.ProjectConfig{
		ProjectName: "cli-test",
		GroupID:     "com.cli.test",
		ArtifactID:  "cli-test",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore"},
		Database:    "postgresql",
	}

	os.Chdir("/Users/arianlc/Documents/Work/trabucotest/test")

	gen, err := generator.New(cfg)
	if err != nil {
		fmt.Printf("Failed: %v\n", err)
		os.Exit(1)
	}

	if err := gen.Generate(); err != nil {
		fmt.Printf("Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Project generated")
}
