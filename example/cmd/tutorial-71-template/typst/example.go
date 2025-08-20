// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package typst

func ExampleData() Report {
	return Report{
		Title:    "Test",
		Subtitle: "Subtest",
		Author: Author{
			Name: "Max Mustermann",
			Birthday: Date{
				Year:  2000,
				Month: 5,
				Day:   1,
			},
			Address: Address{
				Street: "Nordseestr. 1",
				City:   "26131 Oldenburg",
			},
		},
		Company: Company{
			Name: "worldiety GmbH",
			Address: Address{
				Street: "Nordseestr. 2",
				City:   "26131 Oldenburg",
			},
		},
		Trainer: "Muster Maxmann",
		Training: Training{
			Start: Date{
				Year:  2020,
				Month: 8,
				Day:   1,
			},
			End: Date{
				Year:  2023,
				Month: 8,
				Day:   1,
			},
		},
		Entries: []Entry{
			{
				Kind: Day,
				Date: Date{
					Year:  2020,
					Month: 8,
					Day:   1,
				},
				Tasks: []Task{
					{
						Description:       "Lorem Ipsum",
						DurationInMinutes: 30,
					},
					{
						Description:       "Lorem Ipsum",
						DurationInMinutes: 90,
					},
					{
						Description:       "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
						DurationInMinutes: 120,
					},
					{
						Description:       "Lorem Ipsum",
						DurationInMinutes: 30,
					},
				},
				Place: "Betrieb",
			},
			{
				Kind: Day,
				Date: Date{
					Year:  2020,
					Month: 8,
					Day:   2,
				},
				Tasks: []Task{
					{
						Description:       "Lorem Ipsum",
						DurationInMinutes: 30,
					},
					{
						Description:       "Lorem Ipsum",
						DurationInMinutes: 90,
					},
					{
						Description:       "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
						DurationInMinutes: 120,
					},
					{
						Description:       "Lorem Ipsum",
						DurationInMinutes: 30,
					},
				},
				Place: "Betrieb",
			},
			{
				Kind: Signature,
				Date: Date{
					Year:  2020,
					Month: 8,
					Day:   31,
				},
				Tasks: []Task{},
				Place: "Schule",
			},
		},
	}
}
