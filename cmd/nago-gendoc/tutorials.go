// Hugo only has access to files that are located within the hugo project.
// It is therefore necessary that all tutorial files are copied into the hugo project in advance.
package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// generateDocsForTutorials generates Markdown for all tutorials in example/cmd.
func generateDocsForTutorials(
	projectRoot string,
	tutorialsToComponentMap map[string][]string,
	componentToTypeMap map[string]ComponentType,
) error {
	src := filepath.Join(projectRoot, "example/cmd")

	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), "index.md") {
			return nil
		}

		tutorialID := filepath.Base(filepath.Dir(path))

		components, found := tutorialsToComponentMap[tutorialID]
		if !found || len(components) == 0 {
			return removeSeeAlsoSection(path)
		}

		return updateSeeAlsoSection(path, components, componentToTypeMap)
	})

	return err
}

// removeSeeAlsoSection removes the "See also" section of the file.
func removeSeeAlsoSection(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	content := string(data)
	seeAlsoPattern := regexp.MustCompile(`(?s)## See also\n(.*?)(\n## |\n\z)`)
	if seeAlsoPattern.MatchString(content) {
		content = seeAlsoPattern.ReplaceAllString(content, "\n"+"$2")
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// updateSeeAlsoSection searches the file for a "See also" section that contains all linked components.
// If a section  already exist, it is replaced. Otherwise, a new section is appended.
func updateSeeAlsoSection(
	path string,
	components []string,
	componentToTypeMap map[string]ComponentType,
) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	content := string(data)

	newSeeAlsoSection := createMarkdownForSeeAlso(components, componentToTypeMap)

	// find old "See also" section
	seeAlsoPattern := regexp.MustCompile(`(?s)## See also\n(.*?)(\n## |\n\z)`)
	if seeAlsoPattern.MatchString(content) {
		// replace
		content = seeAlsoPattern.ReplaceAllString(content, newSeeAlsoSection+"$2")
	} else {
		// append if there is no "See also" section
		content = strings.TrimSpace(content) + "\n\n" + newSeeAlsoSection + "\n"
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// createMarkdownForSeeAlso generates the "See also" section for tutorials containing all linked components.
func createMarkdownForSeeAlso(
	components []string,
	componentToTypeMap map[string]ComponentType,
) string {
	var sb strings.Builder

	slices.Sort(components)

	sb.WriteString("## See also\n")
	for _, c := range components {
		componentPathInHugo, err := getComponentPathInHugo(c, componentToTypeMap)
		if err != nil {
			slog.Error("Error getting component path in hugo file: ", slog.Any("err", err))
			continue
		}
		sb.WriteString(fmt.Sprintf("- [%s](%s)\n", c, componentPathInHugo))
	}
	return sb.String()
}

// getComponentPathInHugo generates the correct path for a linked component within the hugo project.
// It uses the map to look up the ComponentType and create the path e.g. docs/components/composite/code_editor.
func getComponentPathInHugo(
	component string,
	componentToTypeMap map[string]ComponentType,
) (string, error) {
	componentDirName := strings.ReplaceAll(strings.ToLower(component), " ", "_")

	switch componentToTypeMap[component] {
	case Basic:
		return fmt.Sprintf("../../components/basic/%s", componentDirName), nil
	case Utility:
		return fmt.Sprintf("../../components/utiltity/%s", componentDirName), nil
	case Layout:
		return fmt.Sprintf("../../components/layout/%s", componentDirName), nil
	case FeedbackAndOverlay:
		return fmt.Sprintf("../../components/feedback-and-overlay/%s", componentDirName), nil
	case Composite:
		return fmt.Sprintf("../../components/composite/%s", componentDirName), nil
	default:
		return "", fmt.Errorf("unknown component: %s", component)
	}
}

func copyFilesToHugo(projectRoot string) {
	srcRoot := filepath.Join(projectRoot, "example/cmd")
	dstRoot := filepath.Join(projectRoot, "docs/nago.dev/content/docs/examples")

	if err := copyDir(srcRoot, dstRoot, srcRoot); err != nil {
		log.Fatalf("Fehler beim Kopieren: %v", err)
	}
}

func copyFile(srcFile, dstFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dstFile), os.ModePerm); err != nil {
		return err
	}

	dst, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func copyDir(srcDir, dstDir, rootDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstName := entry.Name()

		// Special case: README.md from the root dir. This will be the start page for all tutorials and must be saved in _index.md
		if dstName == "README.md" && sameDir(srcDir, rootDir) {
			dstName = "_index.md"
		}

		dstPath := filepath.Join(dstDir, dstName)

		if entry.IsDir() {
			// copy dir recursive
			if err := copyDir(srcPath, dstPath, rootDir); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func sameDir(srcDir, rootDir string) bool {
	absA, err := filepath.Abs(srcDir)
	if err != nil {
		return false
	}

	absB, err := filepath.Abs(rootDir)
	if err != nil {
		return false
	}

	return absA == absB
}
