//go:generate go run .

package main

import (
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed tutorial-component-linking.yaml
var linkingFile []byte

func main() {
	projectRoot := findProjectRoot()

	var tutorialToComponentMap map[string][]string
	if err := yaml.Unmarshal(linkingFile, &tutorialToComponentMap); err != nil {
		log.Fatal("Error while parsing tutorial component linking YAML: ", err)
	}

	// Backward mapping
	componentToTutorialMap := map[string][]string{}
	for tut, comps := range tutorialToComponentMap {
		for _, comp := range comps {
			componentToTutorialMap[comp] = append(componentToTutorialMap[comp], tut)
		}
	}

	componentToTypeMap, err := generateDocsForComponents(projectRoot, componentToTutorialMap)
	if err != nil {
		log.Fatal("Error while generating docs for components: ", err)
	}

	err = generateDocsForTutorials(projectRoot, tutorialToComponentMap, componentToTypeMap)
	if err != nil {
		log.Fatal("Error while generating docs for tutorials: ", err)
	}

	copyFilesToHugo(projectRoot)

	cmd := exec.Command("hugo", "build")
	cmd.Dir = filepath.Join(findProjectRoot(), "docs/nago.dev")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Hugo working dir:", cmd.Dir)
	fmt.Println("Starte Hugo Build in docs")
	if err = cmd.Run(); err != nil {
		log.Fatal("Error while running hugo build: ", err)
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
