package main

import (
	"fmt"
	"github.com/worldiety/xtractdoc/domain/api"
	"log/slog"
	"os"
	"path"
	"regexp"
	"strings"
)

type DocSystem struct {
	DisplayName string
	DirName     string
	Type        *api.Type
}

// updateListOfSystems takes all existing systems from Nago and writes them to a Markdown table.
// It also replaces the "List of Systems" section in docs/nago.dev/content/docs/systems_index.md.
func updateListOfSystems(
	projectRoot string,
	systems []DocSystem,
) error {
	filePath := path.Join(projectRoot, "docs/nago.dev/content/docs/systems/_index.md")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	content := string(data)

	re := regexp.MustCompile(`(?s)(## List of Systems\s*\n).*?(\n## |\n\z)`)
	if re.MatchString(content) {
		var tableBuilder strings.Builder
		tableBuilder.WriteString("| System | Description | Details |\n")
		tableBuilder.WriteString("|--------|------------|---------|\n")
		for _, sys := range systems {
			description := formatDocWithBullets(sys.Type.Doc)
			tableBuilder.WriteString(fmt.Sprintf(
				"| `%s` | %s | [Details »](%s) |\n",
				sys.DisplayName,
				description,
				sys.DirName,
			))
		}
		table := tableBuilder.String()
		newContent := re.ReplaceAllString(content, "${1}"+table+"${2}")

		if err = os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("error writing file: %w", err)
		}
	} else {
		slog.Info("No list of Systems found")
	}

	return nil
}

// formatDocWithBullets converts bullet points with "- " into <ul><li></li></ul> as this format is necessary for markdown.
func formatDocWithBullets(doc string) string {
	lines := strings.Split(doc, "\n")

	var intro []string
	var bullets []string
	inBullets := false

	for _, l := range lines {
		trimmed := strings.TrimSpace(l)

		// bullet points start
		if strings.HasPrefix(trimmed, "- ") {
			inBullets = true
			bullets = append(bullets, "<li>"+strings.TrimPrefix(trimmed, "- ")+"</li>")
			continue
		}

		if inBullets {
			if trimmed != "" && strings.HasPrefix(trimmed, "- ") {
				bullets = append(bullets, "<li>"+strings.TrimPrefix(trimmed, "- ")+"</li>")
			}
			continue
		}

		// everything before is intro
		intro = append(intro, trimmed)
	}

	result := strings.Join(intro, " ")
	if len(bullets) > 0 {
		result += "<ul>" + strings.Join(bullets, "") + "</ul>"
	}
	return result
}
