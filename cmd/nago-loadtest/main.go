package main

import (
	"fmt"
	"go.wdy.de/nago/testing"
)

func main() {
	err := testing.NewTester().Test()
	if err != nil {
		fmt.Println(err)
	}
}
