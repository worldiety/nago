// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package main provides a generator for extracting relevant content from the nago code base to create
// an AI agent prompt.
package main

import (
	"fmt"
	"os"

	"github.com/worldiety/option"
	"go.wdy.de/nago/app/genprompt/ast"
)

func main() {
	if err := realMain(); err != nil {
		panic(err)
	}
}

func realMain() error {
	p := ast.NewParser()
	if err := p.Parse(); err != nil {
		return err
	}

	option.MustZero(os.WriteFile("mistral_finetune.jsonl", []byte(p.MistralDataSet()), os.ModePerm))
	//fmt.Println(p.MistralDataSet())
	fmt.Print(p.String())
	return nil
}
