//go:generate go run .

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	projectRoot := findProjectRoot()

	copyFilesToHugo(projectRoot)
	generateDocsForComponents(projectRoot)

	cmd := exec.Command("hugo", "build")
	cmd.Dir = filepath.Join(findProjectRoot(), "docs")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Hugo working dir:", cmd.Dir)
	fmt.Println("Starte Hugo Build in docs")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

// findProjectRoot tries to locate the project root based on known structure (e.g. go.mod location)
func findProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// walk up until go.mod is found
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			log.Fatal("Could not find project root (no go.mod)")
		}
		dir = parent
	}
}
