// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package errs reports errors as values instead of panicking.
package errs

import (
	"errors"

	"go.wdy.de/nago/pkg/gostrict/std/text"
)

type Account struct {
	Balance int32
}

func (a *Account) Withdraw(n int32) error {
	if n <= 0 {
		return errors.New("amount must be positive")
	}
	if n > a.Balance {
		return errors.New("insufficient funds")
	}
	a.Balance -= n
	return nil
}

func Average(sum, count int32) (int32, error) {
	if count == 0 {
		return 0, errors.New("no values")
	}
	return sum / count, nil
}

func Run(amount int32) string {
	acc := &Account{Balance: 100}
	out := ""
	for i := int32(0); i < 3; i++ {
		if err := acc.Withdraw(amount); err != nil {
			out += err.Error() + ";"
			continue
		}
		out += text.FromInt(acc.Balance) + ";"
	}
	avg, err := Average(acc.Balance, amount)
	if err != nil {
		return out + err.Error()
	}
	return out + text.FromInt(avg)
}
