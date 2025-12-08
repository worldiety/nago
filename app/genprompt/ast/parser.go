// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ast

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/worldiety/option"
)

//go:embed prompt_preamble.md
var promptPreamble string

type pkg struct {
	name       string
	importPath string
	content    strings.Builder
}

type example struct {
	quest  string
	answer string
}
type Parser struct {
	packages map[string]*pkg
	modName  string
	modPath  string
	examples []example
}

func NewParser() *Parser {
	return &Parser{packages: make(map[string]*pkg)}
}

func findGoMod(dir string) (string, error) {
	for {
		modPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(modPath); err == nil {
			return modPath, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found")
}

func modulePath(goModPath string) (string, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module line not found")
}

// findProjectRoot tries to locate the project root based on known structure (e.g. go.mod location)
func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// walk up until go.mod is found
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("Could not find project root (no go.mod)")
		}
		dir = parent
	}
}

func (p *Parser) Parse() error {
	p.packages = make(map[string]*pkg)
	root, err := findProjectRoot()
	if err != nil {
		return err
	}

	goModFile, err := findGoMod(root)
	if err != nil {
		return err
	}

	modPath, err := modulePath(goModFile)
	if err != nil {
		return err
	}

	p.modName = modPath
	p.modPath = root

	if err := p.loadExamples(); err != nil {
		return err
	}

	skiplist := []string{
		"/presentation/icons/hero",
		"/presentation/icons/material",
		"/example/cmd",
	}

	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			slog.Info("ignore dir", "file", path)
			return filepath.SkipDir
		}

		rel := strings.TrimPrefix(path, root)
		if d.IsDir() && slices.Contains(skiplist, rel) {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		if err := p.processFile(path); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (p *Parser) processFile(filename string) error {
	fset := token.NewFileSet()

	// Parse including comments
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Walk and remove bodies
	var newDecls []ast.Decl

	for _, decl := range f.Decls {
		keep := true

		switch d := decl.(type) {
		case *ast.FuncDecl:
			if !token.IsExported(d.Name.Name) {
				keep = false
			}
		case *ast.GenDecl:
			// z.B. var, const, type
			var specs []ast.Spec
			for _, s := range d.Specs {
				switch spec := s.(type) {
				case *ast.ValueSpec:
					var newNames []*ast.Ident
					var newValues []ast.Expr
					for i, name := range spec.Names {
						if token.IsExported(name.Name) {
							newNames = append(newNames, name)
							if i < len(spec.Values) {
								newValues = append(newValues, spec.Values[i])
							}
						}
					}
					if len(newNames) > 0 {
						spec.Names = newNames
						spec.Values = newValues
						specs = append(specs, spec)
					}
				case *ast.TypeSpec:
					if token.IsExported(spec.Name.Name) {
						specs = append(specs, spec)
					}
				}
			}
			if len(specs) == 0 {
				keep = false
			} else {
				d.Specs = specs
			}
		}

		if keep {
			newDecls = append(newDecls, decl)
		} else {
			purgeInlineComments(f, decl.Pos(), decl.End())
		}
	}

	for _, n := range newDecls {
		switch fn := n.(type) {
		case *ast.FuncDecl:
			purgeInlineComments(f, fn.Body.Lbrace, fn.Body.Rbrace)
			fn.Body = &ast.BlockStmt{
				Lbrace: fn.Pos(),
				List:   nil,
				Rbrace: fn.Pos(),
			}

		}
	}

	// purge private struct field
	purgePrivateFields(f, newDecls)

	f.Decls = newDecls

	rel, err := filepath.Rel(p.modPath, filepath.Dir(filename))
	if err != nil {
		return err
	}
	importPath := filepath.ToSlash(filepath.Join(p.modName, rel))

	pck := p.packages[importPath]
	if pck == nil {
		pck = new(pkg)
		pck.importPath = importPath
		pck.name = f.Name.Name
		p.packages[importPath] = pck
	}

	// purge license header comment
	i := 0
	for i < len(f.Comments) && f.Comments[i].End() < f.Name.Pos() {
		i++
	}
	f.Comments = f.Comments[i:]

	// Render the code with standard formatting
	var tmpBuf strings.Builder
	cfg := &printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}
	if err := cfg.Fprint(&tmpBuf, fset, f); err != nil {
		return err
	}

	tmpStr := strings.Replace(tmpBuf.String(), "package "+f.Name.Name, "", 1)

	pck.content.WriteString(tmpStr)
	return nil
}

func purgeInlineComments(f *ast.File, lbrace, rbrace token.Pos) {
	var tmp []*ast.CommentGroup
	for _, cg := range f.Comments {
		deleted := cg.Pos() >= lbrace && cg.Pos() <= rbrace
		if deleted {
			continue
		}

		tmp = append(tmp, cg)
	}

	f.Comments = tmp
}

func purgePrivateFields(f *ast.File, decls []ast.Decl) {
	for _, decl := range decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			var newFields []*ast.Field
			for _, field := range structType.Fields.List {
				var newNames []*ast.Ident

				for _, name := range field.Names {
					if token.IsExported(name.Name) {
						newNames = append(newNames, name)
					} else {
						if field.Comment != nil {
							purgeInlineComments(f, field.Comment.Pos(), field.Comment.End())
						}
					}
				}

				if len(newNames) > 0 {
					field.Names = newNames
					newFields = append(newFields, field)
				}
			}

			structType.Fields.List = newFields
		}
	}

}

func (p *Parser) String() string {
	var sb strings.Builder
	sb.WriteString(promptPreamble)

	sb.WriteString("========================================\n\n")
	sb.WriteString("## API signatures (compressed and function bodies omitted)\n\n")
	for _, key := range slices.Sorted(maps.Keys(p.packages)) {
		pck := p.packages[key]

		sb.WriteString("### Package " + pck.importPath + "\n\n")

		sb.WriteString(fmt.Sprintf("```\n// import \"%s\"\npackage %s", pck.importPath, pck.name))
		sb.WriteString(pck.content.String())
		sb.WriteString("\n```\n\n")

	}

	return sb.String()
}

func (p *Parser) typeFromPkg(pkgName string) string {
	var sb strings.Builder
	pck := p.packages[pkgName]

	sb.WriteString("### Package " + pck.importPath + "\n\n")

	sb.WriteString(fmt.Sprintf("```\n// import \"%s\"\npackage %s", pck.importPath, pck.name))
	sb.WriteString(pck.content.String())
	sb.WriteString("\n```\n\n")
	return sb.String()
}

func (p *Parser) splitTypesFromPkg(pkgName string) []string {
	tmp := p.typeFromPkg(pkgName)
	return chunkString(tmp, 25000)
}

func (p *Parser) pkgList() string {
	var sb strings.Builder
	sb.WriteString("Nago provides the following packages:\n")
	for _, key := range slices.Sorted(maps.Keys(p.packages)) {
		sb.WriteString(fmt.Sprintf("- %s\n", key))
	}

	return sb.String()
}

func (p *Parser) loadExamples() error {
	root := filepath.Join(p.modPath, "example", "cmd")
	files, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !(file.IsDir() && strings.HasPrefix(file.Name(), "tutorial")) {
			continue
		}

		fname := filepath.Join(root, file.Name(), "main.go")
		buf, err := os.ReadFile(fname)
		if err != nil {
			return err
		}

		if len(buf) > 25000 {
			slog.Error("file to large as mistral answer", "file", fname)
			continue
		}

		sbuf := string(buf)
		sbuf = strings.TrimPrefix(sbuf, "// Copyright (c) 2025 worldiety GmbH\n//\n// This file is part of the NAGO Low-Code Platform.\n// Licensed under the terms specified in the LICENSE file.\n//\n// SPDX-License-Identifier: Custom-License")

		tokens := strings.SplitN(file.Name(), "-", 3)
		p.examples = append(p.examples, example{
			quest:  "How can I implement " + tokens[len(tokens)-1] + "?",
			answer: sbuf,
		})
	}

	return nil
}

func (p *Parser) MistralDataSet() string {
	const u = "user"
	const a = "assistant"
	var ds []mistralConvTextInstruct
	ds = append(ds, mistralConvTextInstruct{Messages: []mMsg{
		{
			Role:    u,
			Content: "How is the nago framework structured? Which packages are available?",
		},
		{
			Role:    a,
			Content: p.pkgList(),
		},
	}})
	for _, key := range slices.Sorted(maps.Keys(p.packages)) {
		for _, s := range p.splitTypesFromPkg(key) {
			ds = append(ds, mistralConvTextInstruct{Messages: []mMsg{
				{
					Role:    u,
					Content: fmt.Sprintf("Which types are defined in package %s?", key),
				},

				{
					Role:    a,
					Content: s,
				},
			}})
		}

	}

	for _, ex := range p.examples {
		ds = append(ds, mistralConvTextInstruct{
			Messages: []mMsg{
				{
					Role:    u,
					Content: ex.quest,
				},
				{
					Role:    a,
					Content: ex.answer,
				},
			},
		})
	}

	ds = chunkLargeMessages(ds)

	var buf strings.Builder
	for _, d := range ds {
		buf.WriteString(string(option.Must(json.Marshal(d))))
		buf.WriteString("\n")
	}

	return buf.String()
}

func chunkLargeMessages(ds []mistralConvTextInstruct) []mistralConvTextInstruct {
	slices.SortFunc(ds, func(a, b mistralConvTextInstruct) int {
		return -(len(a.Messages[1].Content) - len(b.Messages[1].Content))
	})

	for i := range 10 {
		fmt.Println(ds[i].Messages[0].Content, len(ds[i].Messages[1].Content))
	}

	return ds
}

type mistralConvTextInstruct struct {
	Messages []mMsg `json:"messages"`
}

type mMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func chunkString(s string, maxRunes int) []string {
	if maxRunes <= 0 {
		return nil
	}

	r := []rune(s)
	n := len(r)
	if n == 0 {
		return []string{}
	}

	var out []string
	pos := 0

	for pos < n {
		// Wenn der Rest kürzer ist als maxRunes => letzter Chunk
		if pos+maxRunes >= n {
			out = append(out, string(r[pos:]))
			break
		}

		// Chunk-Kandidat
		end := pos + maxRunes
		chunkRunes := r[pos:end]

		// Im Chunk nach letztem Linebreak suchen
		text := string(chunkRunes)
		lastNL := strings.LastIndex(text, "\n")

		if lastNL == -1 {
			// Kein Linebreak -> hart schneiden
			out = append(out, text)
			pos = end
			continue
		}

		// Linebreak gefunden → schneiden *bis inklusive* '\n'
		cutPos := pos + lastNL + 1 // +1, damit '\n' mitgenommen wird

		out = append(out, string(r[pos:cutPos]))
		pos = cutPos
	}

	return out
}
