//go:generate go run .

package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/worldiety/xtractdoc/domain/api"
	"github.com/worldiety/xtractdoc/domain/app"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type NagoProject struct {
	Components map[string]DocComponent
	Systems    []DocSystem
}

func (np *NagoProject) systemsAlphabetic() {
	sort.Slice(np.Systems, func(i, j int) bool {
		return np.Systems[i].DisplayName < np.Systems[j].DisplayName
	})
}

//go:embed tutorial-component-linking.yaml
var linkingFile []byte

func main() {
	module, err := parseProject()
	if err != nil {
		log.Fatal("Error while parsing project: ", err)
	}

	categorizedNagoProject := categorizeModule(module)
	categorizedNagoProject.systemsAlphabetic()

	projectRoot := findProjectRoot()

	var tutorialToComponentMap map[string][]string
	if err = yaml.Unmarshal(linkingFile, &tutorialToComponentMap); err != nil {
		log.Fatal("Error while parsing tutorial component linking YAML: ", err)
	}

	// Backward mapping
	componentToTutorialMap := map[string][]string{}
	for tut, comps := range tutorialToComponentMap {
		for _, comp := range comps {
			componentToTutorialMap[comp] = append(componentToTutorialMap[comp], tut)
		}
	}

	componentToTypeMap, err := generateDocsForComponents(categorizedNagoProject.Components, projectRoot, componentToTutorialMap)
	if err != nil {
		log.Fatal("Error while generating docs for components: ", err)
	}

	err = generateDocsForTutorials(projectRoot, tutorialToComponentMap, componentToTypeMap)
	if err != nil {
		log.Fatal("Error while generating docs for tutorials: ", err)
	}

	copyFilesToHugo(projectRoot)

	err = updateListOfSystems(projectRoot, categorizedNagoProject.Systems)
	if err != nil {
		log.Fatal("Error while generating docs for systems: ", err)
	}

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

// categorizeModule takes an api.Module and converts it into a NagoProject.
// Therefore, it analyses all api.Package and api.Type and categorizes them to specific types of the NagoProject.
// All api.Type that don't match will be ignored.
func categorizeModule(module api.Module) *NagoProject {
	result := &NagoProject{
		Components: make(map[string]DocComponent),
		Systems:    []DocSystem{},
	}

	componentPattern := regexp.MustCompile(`is (a|an) (basic|layout|utility|feedback|overlay|composite) component\s*\([^)]+\)`)
	systemPattern := regexp.MustCompile(`is a nago system\s*\([^)]+\)`)

	for _, pkg := range module.Packages {
		for name, typ := range pkg.Types {

			// Components
			if componentPattern.MatchString(typ.Doc) {
				addComponent(result, name, typ, componentPattern)
			}

			// Systems
			if systemPattern.MatchString(typ.Doc) {
				addSystem(result, name, typ, systemPattern)
			}
		}
	}

	return result
}

// addSystem creates a DocSystem and adds it to the NagoProject.
// In addition, the api.Type Doc field is cleaned up and the matching pattern is removed.
func addSystem(
	result *NagoProject,
	name string,
	typ *api.Type,
	systemPattern *regexp.Regexp,
) {
	cleanedDoc := cleanDoc(typ.Doc, systemPattern)
	displayName := extractDisplayName(typ.Doc)
	if displayName == "" {
		displayName = name
	}
	result.Systems = append(result.Systems, DocSystem{
		DisplayName: displayName,
		DirName:     strings.ToLower(strings.ReplaceAll(displayName, " ", "_")),
		Type: &api.Type{
			Doc:         cleanedDoc,
			BaseType:    typ.BaseType,
			Stereotypes: typ.Stereotypes,
			Factories:   typ.Factories,
			Methods:     typ.Methods,
			Singletons:  typ.Singletons,
			Fields:      typ.Fields,
			Enumerals:   typ.Enumerals,
		},
	})
}

// addComponent creates a DocComponent and adds it to the NagoProject.
// In addition, the api.Type Doc field is cleaned up and the matching pattern is removed.
func addComponent(
	result *NagoProject,
	name string,
	typ *api.Type,
	componentPattern *regexp.Regexp,
) {
	cleanedDoc := cleanDoc(typ.Doc, componentPattern)
	displayName := extractDisplayName(typ.Doc)
	if displayName == "" {
		displayName = name
	}

	result.Components[name] = DocComponent{
		DisplayName:   displayName,
		Related:       getRelatedTypesInOrder(typ),
		DirName:       strings.ToLower(strings.ReplaceAll(displayName, " ", "_")),
		ComponentType: filterComponentType(typ),
		Type: &api.Type{
			Doc:         cleanedDoc,
			BaseType:    typ.BaseType,
			Stereotypes: typ.Stereotypes,
			Factories:   typ.Factories,
			Methods:     typ.Methods,
			Singletons:  typ.Singletons,
			Fields:      typ.Fields,
			Enumerals:   typ.Enumerals,
		},
	}
}

func parseProject() (api.Module, error) {
	var cfg app.Config
	cfg.Reset()
	cfg.Flags(flag.CommandLine)
	flag.Parse()
	cfg.OutputFormat = "json"

	buf, err := app.Apply(cfg)
	if err != nil {
		return api.Module{}, err
	}

	var module api.Module
	err = json.Unmarshal(buf, &module)
	if err != nil {
		return api.Module{}, err
	}

	return module, nil
}

// cleanDoc removes the assignment record from the comments e.g. TText is a basic component(Text), as this should not be displayed in the docs.
func cleanDoc(doc string, pattern *regexp.Regexp) string {
	parts := strings.Split(doc, ".")
	var cleaned []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if pattern.MatchString(strings.ToLower(part)) {
			continue
		}

		cleaned = append(cleaned, part)
	}

	if len(cleaned) == 0 && !pattern.MatchString(strings.ToLower(doc)) {
		return strings.TrimSpace(doc)
	}

	if len(cleaned) > 0 {
		return strings.Join(cleaned, ". ") + "."
	}

	return ""
}
