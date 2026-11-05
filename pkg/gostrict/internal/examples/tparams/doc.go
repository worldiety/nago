// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

// Package tparams calls the methods of type parameters, whose constraint is an interface with methods: as a method
// value, e.g. f := s.area, and directly. Each other file is a program of its own together with this one.
package tparams

type shape interface{ area() int32 }

type scaler interface{ scale(k int32) int32 }

type sq struct{ n int32 }

func (q sq) area() int32 { return q.n * q.n }

func (q sq) scale(k int32) int32 { return q.n * k }

type rect struct{ w, h int32 }

func (r *rect) area() int32 { return r.w * r.h }
