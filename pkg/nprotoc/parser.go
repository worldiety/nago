package nprotoc

import (
	"bytes"
	"fmt"
	"github.com/worldiety/enum/json"
	"io/fs"
	"maps"
	"path/filepath"
	"slices"
	"strings"
)

func Parse(fsys fs.FS) (map[Typename]Declaration, error) {
	res := map[Typename]Declaration{}
	check := map[int][]string{}

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if d.Type().IsRegular() && strings.HasSuffix(path, ".json") {
			buf, err := fs.ReadFile(fsys, path)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}

			var decl Declaration
			dec := json.NewDecoder(bytes.NewBuffer(buf))
			dec.DisallowUnknownFields()
			if err := dec.Decode(&decl); err != nil {
				//if err := json.Unmarshal(buf, &decl); err != nil {
				return fmt.Errorf("cannot parse '%s': %w", path, err)
			}

			if withId, ok := decl.(IdentityTypeDeclaration); ok {
				id := withId.ID()

				check[id] = append(check[id], path)
			}

			if rec, ok := decl.(Record); ok {
				last := -1
				for _, id := range slices.Sorted(maps.Keys(rec.Fields)) {
					if id < 1 || id > 30 {
						if id > 30 {
							return fmt.Errorf("record id '%d' overflows field id which is not yet implemented: workaround: map fields together into another record to introduce another level of indirection", id)
						}

						return fmt.Errorf("record declaration %s has an invalid field id: %d", path, id)
					}

					if last == -1 {
						last = int(id)
					} else {
						if last != int(id)-1 {
							return fmt.Errorf("record declaration %s has an invalid sequence field id: %d expected %d", path, id, last+1)
						}
						last = int(id)
					}

				}
			}

			typename := Typename(strings.TrimSuffix(path, ".json"))
			optDoc, err := fs.ReadFile(fsys, filepath.Join(filepath.Dir(path), string(typename)+".md"))
			if err == nil {
				switch t := decl.(type) {
				case Enum:
					t.Doc += string(optDoc)
					decl = t
				case Record:
					t.Doc += string(optDoc)
					decl = t
				case Uint:
					t.Doc += string(optDoc)
					decl = t
				}
			}

			res[typename] = decl
		}

		return nil
	})

	nextFree := func() int {
		m := 0
		for id := range check {
			m = max(m, id)
		}
		return m
	}

	for id, files := range check {
		if id <= 0 {
			return nil, fmt.Errorf("declaration id '%d' must be > 0: %s", id, files[0])
		}

		if len(files) > 1 {
			return nil, fmt.Errorf("declaration id '%d' must be unique but was declared in %s and %s (next free id is: %d)", id, files[0], files[1], nextFree()+1)
		}

	}

	return res, err
}
