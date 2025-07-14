package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/worldiety/xtractdoc/domain/api"
	"github.com/worldiety/xtractdoc/domain/app"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type ComponentType int

const (
	Unknown ComponentType = iota
	Basic
	Layout
	Utility
	FeedbackAndOverlay
	Composite
)

// generateDocsForComponents generates markdown for all components in the packages.
// Make sure to include all needed packages in cfg.Packages and separate them with ;
// e.g. "go.wdy.de/nago/presentation/ui;go.wdy.de/nago/presentation/ui/picker"
func generateDocsForComponents(projectRoot string) {
	var cfg app.Config
	cfg.Reset()
	cfg.Flags(flag.CommandLine)
	flag.Parse()
	cfg.OutputFormat = "json"
	cfg.Packages = "go.wdy.de/nago/presentation/ui;go.wdy.de/nago/presentation/ui/colorpicker;go.wdy.de/nago/presentation/ui/alert"

	buf, err := app.Apply(cfg)
	if err != nil {
		panic(err)
	}

	var module api.Module
	err = json.Unmarshal(buf, &module)
	if err != nil {
		panic(err)
	}

	basicComponentOutputPath := filepath.Join(projectRoot, "/docs/content/docs/components/basic")
	layoutComponentOutputPath := filepath.Join(projectRoot, "/docs/content/docs/components/layout")
	utilComponentOutputPath := filepath.Join(projectRoot, "/docs/content/docs/components/utility")
	feedbackAndOverlayComponentOutputPath := filepath.Join(projectRoot, "/docs/content/docs/components/feedback-and-overlay")
	advancedComponentOutputPath := filepath.Join(projectRoot, "/docs/content/docs/components/composite")

	err = os.MkdirAll(basicComponentOutputPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(layoutComponentOutputPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(utilComponentOutputPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(feedbackAndOverlayComponentOutputPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(advancedComponentOutputPath, os.ModePerm)
	if err != nil {
		panic(err)
	}

	docComponents := make(map[string]DocComponent)

	pattern := regexp.MustCompile(`is (a|an) (basic|layout|utility|feedback|overlay|composite) component\([^)]+\)`)

	createDocComponentMapEntries(module, pattern, docComponents)
	createMarkdownAndCopyToHugo(docComponents, basicComponentOutputPath, layoutComponentOutputPath, utilComponentOutputPath, feedbackAndOverlayComponentOutputPath, advancedComponentOutputPath, pattern)
}

// createMarkdownAndCopyToHugo generates markdown for every component and copies it into the hugo project.
func createMarkdownAndCopyToHugo(
	docComponents map[string]DocComponent,
	basicComponentOutputPath string,
	layoutComponentOutputPath string,
	utilComponentOutputPath string,
	feedbackAndOverlayComponentOutputPath string,
	advancedComponentOutputPath string,
	pattern *regexp.Regexp,
) {
	for _, component := range docComponents {
		var dirPath string

		switch component.ComponentType {
		case Basic:
			dirPath = filepath.Join(basicComponentOutputPath, component.DirName)
		case Layout:
			dirPath = filepath.Join(layoutComponentOutputPath, component.DirName)
		case Utility:
			dirPath = filepath.Join(utilComponentOutputPath, component.DirName)
		case FeedbackAndOverlay:
			dirPath = filepath.Join(feedbackAndOverlayComponentOutputPath, component.DirName)
		case Composite:
			dirPath = filepath.Join(advancedComponentOutputPath, component.DirName)
		default:
			slog.Warn("Unknown component type", "type", component.ComponentType, "Component directory", component.DirName, "Component display name", component.DisplayName)
			continue
		}

		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			panic(err)
		}

		filePath := filepath.Join(dirPath, "_index.md")
		md := createMarkdownForComponent(component, docComponents, pattern)

		err = os.WriteFile(filePath, []byte(md), 0644)
		if err != nil {
			panic(err)
		}
	}
}

// createDocComponentMapEntries creates a map entry for every component.
// Afterward, the display name, file name & related components are known for all components.
func createDocComponentMapEntries(
	module api.Module,
	pattern *regexp.Regexp,
	docComponents map[string]DocComponent,
) {
	for _, pkg := range module.Packages {
		for name, typ := range pkg.Types {
			if !pattern.MatchString(typ.Doc) {
				continue
			}

			related := getRelatedTypesInOrder(typ)
			componentType := filterComponentType(typ)

			displayName := extractDisplayName(typ.Doc)
			if displayName == "" {
				displayName = name
			}

			docComponents[name] = DocComponent{
				DisplayName:   displayName,
				Related:       related,
				DirName:       strings.ToLower(strings.ReplaceAll(displayName, " ", "_")),
				ComponentType: componentType,
				Type:          typ,
			}
		}
	}
}

func filterComponentType(t *api.Type) ComponentType {
	switch {
	case strings.Contains(t.Doc, "is a basic component"):
		return Basic
	case strings.Contains(t.Doc, "is a layout component"):
		return Layout
	case strings.Contains(t.Doc, "is an utility component"):
		return Utility
	case strings.Contains(t.Doc, "is a feedback component") || strings.Contains(t.Doc, "is an overlay component"):
		return FeedbackAndOverlay
	case strings.Contains(t.Doc, "is a composite component"):
		return Composite
	default:
		return Unknown
	}
}

// extractDisplayName extracts the name of a component from the comment.
// e.g. TText -> Text.
func extractDisplayName(doc string) string {
	start := strings.Index(doc, "(")
	end := strings.Index(doc, ")")
	if start >= 0 && end > start {
		return strings.TrimSpace(doc[start+1 : end])
	}
	return ""
}

func createMarkdownForFactory(name string, factory *api.Func) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("### %s\n", name))

	if factory.Doc != "" {
		sb.WriteString(factory.Doc)
	}

	if len(factory.Examples) > 0 {
		for _, ex := range factory.Examples {
			code := strings.Trim(ex.Code, "{}\n")

			createMarkdownForExample(code, ex.Doc, "go", &sb)
		}
	}

	if len(factory.ExecutableExamples) > 0 {
		for _, ex := range factory.ExecutableExamples {
			createMarkdownForExample(ex.Code, ex.Doc, "go", &sb)
		}
	}

	return sb.String()
}

func createMarkdownForExample(code, doc, language string, sb *strings.Builder) {
	createMarkdownForCodeBlock(language, code, sb)

	var exampleImageSrcMatch []string

	if doc != "" {
		exampleImgRe := regexp.MustCompile(`\[(.*?)]`)
		exampleImageSrcMatch = exampleImgRe.FindStringSubmatch(doc)
	}

	if len(exampleImageSrcMatch) > 1 {
		createMarkdownForImage(exampleImageSrcMatch[1], sb)
	}
}

func createMarkdownForCodeBlock(language, code string, sb *strings.Builder) {
	sb.WriteString(fmt.Sprintf("```%s\n", language))
	sb.WriteString(code)
	sb.WriteString("\n```\n")
}

func createMarkdownForImage(path string, sb *strings.Builder) {
	sb.WriteString("\n![](")
	sb.WriteString(path)
	sb.WriteString(")\n")
}

func createMarkdownForComponent(
	component DocComponent,
	docComponents map[string]DocComponent,
	pattern *regexp.Regexp,
) string {
	var sb strings.Builder

	// Front Matter
	sb.WriteString("---\n")
	sb.WriteString("# Content is auto generated\n")
	sb.WriteString("# Manual changes will be overwritten!\n")
	sb.WriteString(fmt.Sprintf("title: %s\n", component.DisplayName))
	sb.WriteString("---\n")

	// Description
	sb.WriteString(cleanDoc(component.Type.Doc, pattern) + "\n")
	sb.WriteString("\n")

	// Constructors
	createMarkdownForConstructors(component, &sb)

	// Methods
	createMarkdownForMethods(component, &sb)

	// Related
	createMarkdownForRelatedComponents(component, docComponents, &sb)

	return sb.String()
}

func createMarkdownForConstructors(component DocComponent, sb *strings.Builder) {
	var constructors []string

	if len(component.Type.Factories) > 0 {
		keys := sortMapKeys(component.Type.Factories)

		for _, k := range keys {
			v := component.Type.Factories[k]

			if unicode.IsUpper(rune(k[0])) {
				md := createMarkdownForFactory(k, v)
				constructors = append(constructors, md)
			}
		}
	}

	if len(constructors) > 0 {
		sb.WriteString("## Constructors\n")

		for _, constructor := range constructors {
			sb.WriteString(constructor)
			sb.WriteString("\n")
		}

		sb.WriteString("---\n")
	}
}

func createMarkdownForMethods(
	component DocComponent,
	sb *strings.Builder,
) {
	// Render an ora should not be shown in the docs
	delete(component.Type.Methods, "Render")
	delete(component.Type.Methods, "ora")

	if len(component.Type.Methods) > 0 {
		sb.WriteString("## Methods\n")
		sb.WriteString("| Method | Description |\n")
		sb.WriteString("|--------| ------------|\n")

		sortedMethods := sortMapKeys(component.Type.Methods)
		for _, key := range sortedMethods {
			value := component.Type.Methods[key]

			sb.WriteString(fmt.Sprintf("| `%s(", key))

			var counter int

			for paramName, param := range value.Params {
				if counter > 0 {
					sb.WriteString(", ")
				}

				sb.WriteString(fmt.Sprintf("%s %s", paramName, param.BaseType))
				counter++
			}

			cleanedDoc := strings.ReplaceAll(value.Doc, "\n", " ")
			cleanedDoc = strings.TrimRightFunc(cleanedDoc, unicode.IsSpace)

			sb.WriteString(fmt.Sprintf(")` | %s |\n", cleanedDoc))
		}

		sb.WriteString("---\n")
	}
}

// createMarkdownForRelatedComponents creates a markdown section
func createMarkdownForRelatedComponents(
	component DocComponent,
	docComponents map[string]DocComponent,
	sb *strings.Builder,
) {
	var related []string

	if len(component.Related) > 0 {
		for _, s := range component.Related {
			relatedComponent, ok := docComponents[s]
			if !ok {
				continue
			}

			switch relatedComponent.ComponentType {
			case Basic:
				related = append(related, fmt.Sprintf("- [%s](%s)\n", relatedComponent.DisplayName, "../../basic/"+relatedComponent.DirName+"/"))
			case Layout:
				related = append(related, fmt.Sprintf("- [%s](%s)\n", relatedComponent.DisplayName, "../../layout/"+relatedComponent.DirName+"/"))
			case Utility:
				related = append(related, fmt.Sprintf("- [%s](%s)\n", relatedComponent.DisplayName, "../../utility/"+relatedComponent.DirName+"/"))
			case FeedbackAndOverlay:
				related = append(related, fmt.Sprintf("- [%s](%s)\n", relatedComponent.DisplayName, "../../feedback-and-overlay/"+relatedComponent.DirName+"/"))
			case Composite:
				related = append(related, fmt.Sprintf("- [%s](%s)\n", relatedComponent.DisplayName, "../../composite/"+relatedComponent.DirName+"/"))
			default:
				slog.Warn("Unknown component type. Could not add to related types", "type", relatedComponent.ComponentType)
			}
		}
	}

	if len(related) > 0 {
		sb.WriteString("## Related\n\n")

		for _, relatedComponent := range related {
			sb.WriteString(relatedComponent)
		}
	}
}

func sortMapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

// cleanDoc removes the assignment record of a component from the comments, as this should not be displayed in the docs.
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

type DocComponent struct {
	DisplayName   string
	Related       []string
	DirName       string
	ComponentType ComponentType
	Type          *api.Type
}

// getRelatedTypesInOrder collects all parameter and result types for every method of the *api.Type.
// Using the map prevents collecting duplicates and reduce memory usage.
func getRelatedTypesInOrder(typ *api.Type) []string {
	typeSet := make(map[string]struct{})
	var relatedTypes []string

	for _, value := range typ.Methods {
		for _, parameter := range value.Params {
			if parameter.BaseType == typ.BaseType {
				continue
			}

			if _, exist := typeSet[parameter.BaseType]; !exist {
				typeSet[parameter.BaseType] = struct{}{}
				relatedTypes = append(relatedTypes, parameter.BaseType)
			}
		}

		for _, result := range value.Results {
			if result.BaseType == typ.BaseType {
				continue
			}

			if _, exist := typeSet[result.BaseType]; !exist {
				typeSet[result.BaseType] = struct{}{}
				relatedTypes = append(relatedTypes, result.BaseType)
			}
		}
	}

	sort.Strings(relatedTypes)

	return relatedTypes
}
