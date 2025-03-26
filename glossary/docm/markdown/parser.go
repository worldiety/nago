// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package markdown

import (
	"go.wdy.de/nago/glossary/docm"
	"regexp"
	"strings"
)

var regExMarkdownImg = regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`)

// trim removes any known garbage patterns.
func trim(s string) string {
	return strings.TrimSpace(s)
}

// Parse understands only a minor subset of markdown, usually only what is needed to
// render godoc properly.
func Parse(text string) docm.Element {
	var seq docm.Sequence

	for _, line := range strings.Split(text, "\n") {
		if trim(line) == "" {

			last, _ := seq.Last()
			switch last.(type) {
			case *docm.Par:
			// do nothing
			case *docm.Heading:
			// do nothing
			default:
				seq = append(seq, &docm.Par{})
			}

		} else if strings.HasPrefix(line, "# ") {
			seq = append(seq, &docm.Heading{Level: 1, Body: &docm.Text{Value: line[2:]}})

		} else if strings.HasPrefix(line, "## ") {
			seq = append(seq, &docm.Heading{Level: 1, Body: &docm.Text{Value: line[3:]}})

		} else if strings.HasPrefix(line, "  - ") {
			last, _ := seq.Last()
			switch last := last.(type) {
			case *docm.List:
				last.Add(&docm.Text{Value: line[4:]})
				continue
			default:
				list := &docm.List{}
				list.Add(&docm.Text{Value: line[4:]})
				seq = append(seq, list)
			}

		} else if strings.HasPrefix(line, "\t") {
			last, _ := seq.Last()

			switch last := last.(type) {
			case *docm.Pre:
				last.Add(line)
			default:
				list := &docm.Pre{}
				list.Add(line)
				seq = append(seq, list)
			}
		} else {

			addText := func(line string) {
				last, _ := seq.Last()
				switch last := last.(type) {
				case *docm.Par:
					last.Add(&docm.Text{Value: trim(line) + " "})
				default:
					list := &docm.Par{}
					list.Add(&docm.Text{Value: trim(line) + " "})
					seq = append(seq, list)
				}
			}

			allImages := regExMarkdownImg.FindAllString(line, -1)
			if len(allImages) > 0 {
				for idx, line := range regExMarkdownImg.Split(line, -1) {
					addText(line)
					if idx < len(allImages) {
						seq = append(seq, &docm.Image{URL: urlFromMD(allImages[idx])})
					}
				}
			} else {
				addText(line)
			}

		}
	}

	return seq
}

// something like ![alt text](https://cdn.prod.website-files.com/65c9e1a09853d67c47d4320d/66759ca28472aeb9389f94b5_worldiety-team.jpg)
func urlFromMD(text string) string {
	sol := strings.Index(text, "(")
	return text[sol+1 : len(text)-1]
}
