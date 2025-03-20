package template

import (
	"context"
	"github.com/worldiety/option"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"golang.org/x/text/language"
	"io"
	"io/fs"
	"iter"
	"sync"
)

type ExecType int

func (t ExecType) String() string {
	switch t {
	case Unprocessed:
		return "Unprocessed"
	case TreeTemplatePlain:
		return "TreeTemplatePlain"
	case TreeTemplateHTML:
		return "TreeTemplateHTML"
	case TypstPDF:
		return "TypstPDF"
	case LatexPDF:
		return "LatexPDF"
	case AsciidocPDF:
		return "AsciidocPDF"
	default:
		return "Unknown"
	}
}

const (
	TagPDF  Tag = "pdf"
	TagHTML Tag = "html"
	TagMail Tag = "mail"
)

const (
	Unprocessed       ExecType = iota // return type is always application/zip
	TreeTemplatePlain                 // return type is text/plain
	TreeTemplateHTML                  // text/html
	TypstPDF                          // return type is always application/pdf
	LatexPDF                          // return type is always application/pdf
	AsciidocPDF                       // return type is always application/pdf
)

type ExecOptions struct {
	// Context is used to cancel an execution, e.g. because it takes to long to complete.
	Context context.Context

	// Language represents an already parsed tag, like in [auth.Subject.Language].
	// It may be [language.Und] which is the zero value, to disable any localization effects.
	// To localize a template, you have to put the required files into the magic folder locales/<BCP47-Tag>/...
	Language language.Tag // if zero (language.Und), does not have any effect.

	// TemplateName behaviour is quite complex:
	//  - if template is a [TreeTemplate], parse all tpl files into a single template tree and execute only the named
	//    template. If name is empty, continue without a name.
	//    The result is always a single file. This is useful, if you have different variants and template
	//    subsets, e.g. when rendering static text files. If at least a single *.gohtml file is found,
	//    the html template engine is used to automatically apply escaping and protect against injection attacks.
	//  - All *PDF templates must be executed with an empty template name.
	//    These are large document structures with its own include mechanic, like latex or typst templates, eventually
	//    with templated graphic files (like SVG). These files are also renamed by removing the .tpl suffix.
	TemplateName DefinedTemplateName

	// Model may be nil or whatever fits the template. To know more, inspect the [Project.Examples] of what may
	// be allowed.
	Model any
}

// Execute takes the project and performs the following steps:
//   - load all files from FileSet into memory
//   - copy all localized files on top of it, if tag is not empty (file extensions are .nago.en-US.tpl
//   - if template name is empty, execute each file marked as [File.IsTemplate] through the Go template engine and replace the original file in-memory
//   - if template name is not empty, load all files marked as [File.IsTemplate] at once and execute the defined template by name. Add the result as a new file.
//   - if ExecType is Unprocessed|TextTemplateToText|HtmlTemplateToHtml and there is only a single file result, just return those bytes. Otherwise, return a zip file.
//   - if ExecType is xToPDF try first to render locally and otherwise create a zip file, lookup a secret and try to render using a REST Service.
type Execute func(subject auth.Subject, id ID, options ExecOptions) (io.ReadCloser, error)

// FindAll returns those Project entries, which are allowed by [Project.ReadableBy] (empty means, read by all). Also, the subject must
// have the permission per se. If tags are empty, the filter is ignored. Otherwise, all tags are evaluated using AND
// semantics.
type FindAll func(subject auth.Subject, tags []Tag) iter.Seq2[Project, error]

type FindByID func(subject auth.Subject, id ID) (option.Opt[Project], error)

// Commit overwrites the given project entry with a new version, if allowed. See [Project.WriteableBy] and [PermCommit].
// Note, that Commit will fail, if there are any groups listed, to which the user does not belong, to mitigate
// potential security issues, like inserting malicious templates into otherwise unreachable circles.
type Commit func(subject auth.Subject, project Project) error

// Create inserts the given project and returns the ID which is autogenerated, if empty. Requires [PermCreate].
type Create func(subject auth.Subject, project Project) (ID, error)

type Versions func(subject auth.Subject, id ID) iter.Seq2[VersionID, error]
type FindVersion func(subject auth.Subject, id VersionID) (std.Option[Project], error)

type LoadProjectBlob func(subject auth.Subject, pid ID, file BlobID) (std.Option[io.ReadCloser], error)
type UpdateProjectBlob func(subject auth.Subject, pid ID, filename string, reader io.Reader) error
type DeleteProjectBlob func(subject auth.Subject, pid ID, filename string) error
type CreateProjectBlob func(subject auth.Subject, pid ID, filename string, reader io.Reader) error

type AddRunConfiguration func(subject auth.Subject, pid ID, configuration RunConfiguration) error

type RemoveRunConfiguration func(subject auth.Subject, pid ID, nameOrId string) error

type NewProjectData struct {
	ID          ID
	Name        string
	Description string
	ExecType    ExecType
	Tags        []Tag
	Files       fs.FS
}

// EnsureBuildIn writes the given project data if no such project already exist. Otherwise, it does nothing.
// If force is true, the project will be overwritten, just like as it has been created.
type EnsureBuildIn func(subject auth.Subject, project NewProjectData, force bool) error

type VersionID string
type ID string

type LanguageTag string

type JSONString = string

type DefinedTemplateName = string

type Repository data.Repository[Project, ID]

type UseCases struct {
	FindAll                FindAll
	Execute                Execute
	Create                 Create
	EnsureBuildIn          EnsureBuildIn
	FindByID               FindByID
	LoadProjectBlob        LoadProjectBlob
	UpdateProjectBlob      UpdateProjectBlob
	AddRunConfiguration    AddRunConfiguration
	RemoveRunConfiguration RemoveRunConfiguration
	DeleteProjectBlob      DeleteProjectBlob
	CreateProjectBlob      CreateProjectBlob
}

func NewUseCases(files blob.Store, repository Repository) UseCases {
	var mutex sync.Mutex

	executeFn := NewExecute(files, repository)
	findAllFn := NewFindAll(repository)
	createFn := NewCreate(&mutex, repository)
	ensureBuildInFn := NewEnsureBuildIn(&mutex, repository, files)

	return UseCases{
		FindAll:                findAllFn,
		Execute:                executeFn,
		Create:                 createFn,
		EnsureBuildIn:          ensureBuildInFn,
		FindByID:               NewFindByID(repository),
		LoadProjectBlob:        NewLoadProjectBlob(files, repository),
		UpdateProjectBlob:      NewUpdateProjectBlob(&mutex, files, repository),
		AddRunConfiguration:    NewAddRunConfiguration(&mutex, repository),
		RemoveRunConfiguration: NewRemoveRunConfiguration(&mutex, repository),
		DeleteProjectBlob:      NewDeleteProjectBlob(&mutex, files, repository),
		CreateProjectBlob:      NewCreateProjectBlob(&mutex, files, repository),
	}
}
