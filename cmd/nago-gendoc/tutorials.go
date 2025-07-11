// Hugo only has access to files that are located within the hugo project.
// It is therefore necessary that all tutorial files are copied into the hugo project in advance.
package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func copyFilesToHugo(projectRoot string) {
	srcRoot := filepath.Join(projectRoot, "example/cmd")
	dstRoot := filepath.Join(projectRoot, "docs/content/docs/examples")

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
