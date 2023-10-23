package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	root, _ := os.Getwd()
	fmt.Println(root)
	if err := emitEnums(filepath.Join(root, "container/enum"), 2, 9); err != nil {
		panic(err)
	}
}
