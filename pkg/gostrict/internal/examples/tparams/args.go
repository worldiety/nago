// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package tparams

func argsScale[S scaler](s S, k int32) int32 {
	f := s.scale
	return f(k)
}

// Args calls the method value of a type parameter with an argument.
func Args(n int32) int32 {
	return argsScale(sq{n: n}, 3)
}
