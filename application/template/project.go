package template

import (
	"go.wdy.de/nago/presentation/core"
	"golang.org/x/text/language"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"time"
)

type RunConfiguration struct {
	ID       string
	Name     string
	Template DefinedTemplateName
	Language string
	Model    JSONString
}

type Tag string
type Project struct {
	ID                ID                 `json:"id,omitempty"`
	Name              string             `json:"name,omitempty"`
	Description       string             `json:"description,omitempty"`
	Logo              core.URI           `json:"logo,omitempty"`
	Type              ExecType           `json:"type,omitempty"` // type of evaluation/engine
	Files             []File             `json:"files,omitempty"`
	RunConfigurations []RunConfiguration `json:"runConfigurations,omitempty"`
	Protected         bool               `json:"protected,omitempty"` // just an extra layer of security, for very important templates, like system mail templates
	Tags              []Tag              `json:"tags,omitempty"`      // some arbitrary tags for filtering and inspection
}

func (p Project) Identity() ID {
	return p.ID
}

// Localize applies the localization logic. If a locales folder exists, match against those files and blend them
// onto the default file set and return that. An undefined tag or non locales at all will just return [Project.Default].
func (p Project) Localize(tag language.Tag) []File {
	files := p.Default()
	if tag == language.Und {
		return files
	}

	available := p.Locales()
	if len(available) == 0 {
		return files
	}

	tags := make([]language.Tag, 0, len(available))
	for _, loc := range available {
		tags = append(tags, loc.Tag)
	}
	matcher := language.NewMatcher(tags)
	_, idx, _ := matcher.Match(tag) // at worst, just the first entry is returned
	tags[idx] = tag

	loc := available[idx]

	for _, file := range p.Files {
		if !strings.HasPrefix(file.Filename, loc.Prefix) {
			continue
		}

		targetPath := strings.TrimPrefix(file.Filename, loc.Prefix)
		files = slices.DeleteFunc(files, func(file File) bool {
			return file.Filename == targetPath
		})

		files = append(files, File{
			Filename: targetPath,
			Blob:     file.Blob,
			LastMod:  file.LastMod,
		})
	}

	return files
}

type LocalizedPrefix struct {
	Prefix Filename
	Tag    language.Tag
}

// Locales returns all available language tags in sorted order.
func (p Project) Locales() []LocalizedPrefix {
	tmp := map[string]bool{}
	for _, file := range p.Files {
		if !strings.HasPrefix(file.Filename, "locales") {
			continue
		}

		segments := strings.Split(file.Filename, "/")
		if len(segments) <= 2 {
			// this must be something broken, like a file in locales
			slog.Error("template project contains stale file in locales", "file", file.Filename, "id", p.ID, "name", p.Name)
			continue
		}

		tmp[segments[1]] = true
	}

	res := make([]LocalizedPrefix, 0, len(tmp))
	// todo this sorting is stable but non-sense: start with en, de, fr, es languages first and then append sorted rest
	for _, tagName := range slices.Sorted(maps.Keys(tmp)) {
		tag, err := language.Parse(tagName)
		if err != nil {
			slog.Error("template project contains invalid BCP47 tag in locales", "tag", tagName, "id", p.ID, "name", p.Name)
			continue
		}

		res = append(res, LocalizedPrefix{
			Prefix: "locales/" + tagName,
			Tag:    tag,
		})
	}

	return res
}

// Default returns the default file set.
func (p Project) Default() []File {
	files := make([]File, 0, len(p.Files))
	for _, file := range p.Files {
		if strings.HasPrefix(file.Filename, "locales") {
			continue
		}

		files = append(files, file)
	}

	return files
}

type Filename = string
type BlobID = string

type FileSet map[Filename]File

type File struct {
	// including the path, e.g. index.gohtml or locales/en/index.gohtml. The path rules follow the official fs
	// guidelines, thus starting with / or containing . or .. is invalid.
	Filename Filename

	Blob    BlobID
	LastMod time.Time
}

// IsTemplate inspects the file name
func (f File) IsTemplate() bool {
	return strings.HasSuffix(f.Filename, ".gohtml") || strings.HasSuffix(f.Filename, ".tpl")
}

// Name returns the target (stripped) filename. E.g. index.gohtml becomes index.html or index.html.tpl
// becomes index.html.
func cleanName(name string) string {
	if strings.HasSuffix(name, ".gohtml") {
		return name[:len(name)-7] + ".html"
	}

	if strings.HasSuffix(name, ".tpl") {
		return name[:len(name)-4]
	}

	return name
}
