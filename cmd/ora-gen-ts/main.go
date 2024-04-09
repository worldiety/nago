package main

import (
	"bytes"
	"flag"
	"go.wdy.de/nago/presentation/protocol"
	"log/slog"
	"os"
	"path/filepath"
	"unicode"
)

func main() {
	dir, _ := os.Getwd()
	dir = filepath.Join(dir, "web", "vuejs", "src", "shared", "protocol", "gen")
	outDir := flag.String("output-dir", dir, "the target directory to overwrite files into")
	flag.Parse()

	slog.Info("generating files in", slog.String("dir", *outDir))

	writeFile(*outDir, "interface",
		// emit events
		NewInterface(protocol.EventsAggregated{}),
		NewInterface(protocol.Acknowledged{}),
		NewInterface(protocol.NewComponentRequested{}),
		NewInterface(protocol.ComponentInvalidated{}),

		//emit components
		NewInterface(protocol.Button{}),
	)

	writeFile(*outDir, "union",
		NewUnion("Component", protocol.Components),
		NewUnion("Event", protocol.Events),
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
