//go:generate go run go.wdy.de/nago/cmd/nprotoc -target=../../presentation/proto -source=../../presentation/proto/spec
//go:generate go run go.wdy.de/nago/cmd/nprotoc -lang=ts -target=../../web/vuejs/src/shared/proto -source=../../presentation/proto/spec
//go:generate sh npm_format.sh

package main

import (
	"flag"
	"fmt"
	"go.wdy.de/nago/pkg/nprotoc"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := realMain(); err != nil {
		log.Fatal(err)
	}

}

func realMain() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine working directory")
	}

	specDir := flag.String("source", filepath.Join(dir, "specs"), "specs source directory")
	targetDir := flag.String("target", filepath.Join(dir, "tmp"), "target directory")

	codeTypeGen := flag.String("lang", "go", "one of go|ts")
	flag.Parse()

	if !strings.HasPrefix(*targetDir, "/") {
		*targetDir = filepath.Join(dir, *targetDir)
	}

	if !strings.HasPrefix(*specDir, "/") {
		*specDir = filepath.Join(dir, *specDir)
	}

	declr, err := nprotoc.Parse(os.DirFS(*specDir))
	if err != nil {
		return fmt.Errorf("could not parse specs directory: %w", err)
	}

	compiler := nprotoc.NewCompiler(declr)
	switch *codeTypeGen {
	case "go":
		buf, err := compiler.GenerateGo()
		if err != nil {
			fmt.Println(string(buf))
			return err
		}

		fname := filepath.Join(*targetDir, "protonc_gen.go")

		slog.Info("writing protonc", "file", fname)
		return os.WriteFile(fname, buf, 0644)

	case "ts":
		buf, err := compiler.GenerateTS()
		if err != nil {
			fmt.Println(string(buf))
			return err
		}

		fname := filepath.Join(*targetDir, "nprotoc_gen.ts")

		slog.Info("writing nprotoc", "file", fname)
		return os.WriteFile(fname, buf, 0644)
	default:
		return fmt.Errorf("unknown code type generator %q", *codeTypeGen)
	}
}
