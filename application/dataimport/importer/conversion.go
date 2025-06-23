// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package importer

import (
	"encoding/json"
	"github.com/worldiety/jsonptr"
)

func ToMatch(v any, score float64) (Match, error) {
	obj, err := ToJSON(v)
	if err != nil {
		return Match{}, err
	}

	return Match{Obj: obj, Score: score}, nil
}

func ToJSON(v any) (*jsonptr.Obj, error) {
	if obj, ok := v.(*jsonptr.Obj); ok {
		return obj, nil
	}

	var tmp *jsonptr.Obj
	buf, err := json.Marshal(v)
	if err != nil {
		return tmp, err
	}

	if err := json.Unmarshal(buf, &tmp); err != nil {
		return tmp, err
	}

	return tmp, nil
}

func FromJSON[T any](v *jsonptr.Obj) (T, error) {
	var tmp T
	buf, err := json.Marshal(v)
	if err != nil {
		return tmp, err
	}

	if err := json.Unmarshal(buf, &tmp); err != nil {
		return tmp, err
	}

	return tmp, nil
}
