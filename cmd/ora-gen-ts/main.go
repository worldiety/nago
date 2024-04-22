package main

import (
	"bytes"
	"flag"
	"go.wdy.de/nago/presentation/ora"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"unicode"
)

func main() {
	dir, _ := os.Getwd()
	dir = filepath.Join(dir, "web", "vuejs", "src", "shared", "protocol", "gen")
	outDir := flag.String("output-dir", dir, "the target directory to overwrite files into")
	flag.Parse()

	slog.Info("generating files in", slog.String("dir", *outDir))

	writeFile(*outDir, "interface",
		NewInterface(ora.Themes{}),
		NewInterface(ora.Theme{}),
		NewInterface(ora.Colors{}),
		NewInterface(ora.Resources{}),
	)

	types := append(ora.Events)
	types = append(types, ora.Components...)
	var ifaces []NamedType

	for _, r := range types {
		ifaces = append(ifaces, NewInterface(reflect.New(r).Elem().Interface()))
	}

	writeFile(*outDir, "interface", ifaces...)

	writeFile(*outDir, "union",
		NewUnion("Component", ora.Components),
		NewUnion("Event", ora.Events),
	)
}

func writeFile(dir string, tplName string, nameds ...NamedType) {
	for _, named := range nameds {
		var tmp bytes.Buffer
		err := parsedTemplates.ExecuteTemplate(&tmp, tplName, named)
		if err != nil {
			panic(err)
		}

		fname := filepath.Join(dir, toLowerFirstChar(named.GetName())+".ts")
		if err := os.WriteFile(fname, tmp.Bytes(), os.ModePerm); err != nil {
			panic(err)
		}
	}

}

func toLowerFirstChar(s string) string {
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}
