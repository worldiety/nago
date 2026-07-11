// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package ndbinspector

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/pkg/ndb/tsdb"
	"go.wdy.de/nago/pkg/timeseries"
)

// seriesDB is the accessor a tsdb-backed engine exposes to reach its concrete
// database (see tsdb/engine.go). We type-assert to it rather than importing the
// unexported engine type.
type seriesDB interface {
	DB() *tsdb.DB
}

// ColumnInfo is the metadata-only description of one tsdb column.
type ColumnInfo struct {
	Bucket   string
	Column   string
	Scheme   tsdb.Scheme
	Decimals uint8
	Chunks   int
	Bytes    int64
	// MinMillis / MaxMillis bound the data (inclusive, unix millis). Valid only
	// when HasData is true.
	MinMillis int64
	MaxMillis int64
	HasData   bool
}

// Key is the "bucket/column" identifier used to reference the column in the UI.
func (c ColumnInfo) Key() string { return c.Bucket + "/" + c.Column }

func (c ColumnInfo) Identity() string { return c.Key() }

// Numeric reports whether the column is chartable (SchemeDecimal).
func (c ColumnInfo) Numeric() bool { return c.Scheme == tsdb.SchemeDecimal }

// SeriesPoint is one downsampled chart point.
type SeriesPoint struct {
	Millis int64
	Value  float64
}

// StringRow is one (time, string) value of an enum/string column.
type StringRow struct {
	Millis int64
	Value  string
}

// SeriesRequest bounds an M4-downsampled read of a numeric column. Width is the
// target number of pixel buckets; M4 yields at most 4*Width points regardless of
// how many raw points exist.
type SeriesRequest struct {
	Instance  string
	Engine    string
	Bucket    string
	Column    string
	MinMillis int64
	MaxMillis int64
	Width     int
}

// StringWindowRequest bounds a windowed read of a string/enum column.
type StringWindowRequest struct {
	Instance  string
	Engine    string
	Bucket    string
	Column    string
	MinMillis int64
	MaxMillis int64
	Limit     int // 0 -> DefaultWindowLimit
}

// maxChartWidth caps the M4 pixel width so a single request can never produce an
// unbounded chart series.
const maxChartWidth = 2000

// SeriesEngines lists the tsdb engine instances of one ndb database.
func (uc UseCases) SeriesEngines(subject auth.Subject, instancePath string) ([]EngineRef, error) {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return nil, err
	}
	db, err := uc.dbByPath(instancePath)
	if err != nil {
		return nil, err
	}
	var out []EngineRef
	for info, err := range db.Engines() {
		if err != nil {
			return nil, err
		}
		if info.Kind != tsdb.EngineKind {
			continue
		}
		out = append(out, EngineRef{Instance: instancePath, Name: info.Name, Kind: info.Kind})
	}
	slices.SortFunc(out, func(a, b EngineRef) int { return cmpStr(a.Name, b.Name) })
	return out, nil
}

// series resolves the concrete tsdb DB for the named engine in an instance.
func (uc UseCases) series(instancePath, engine string) (*tsdb.DB, error) {
	db, err := uc.dbByPath(instancePath)
	if err != nil {
		return nil, err
	}
	opt, err := db.LookupEngine(engine)
	if err != nil {
		return nil, err
	}
	if opt.IsNone() {
		return nil, fmt.Errorf("engine %q not found", engine)
	}
	acc, ok := opt.Unwrap().(seriesDB)
	if !ok {
		return nil, fmt.Errorf("engine %q is not a tsdb engine", engine)
	}
	return acc.DB(), nil
}

// column resolves a "bucket/column" pair to an open *tsdb.Column.
func (uc UseCases) column(instancePath, engine, bucket, column string) (*tsdb.Column, error) {
	db, err := uc.series(instancePath, engine)
	if err != nil {
		return nil, err
	}
	col, ok, err := db.LookupColumn(bucket, column)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("column %q/%q not found", bucket, column)
	}
	return col, nil
}

// Columns lists the columns of a tsdb engine with cheap, metadata-only stats.
func (uc UseCases) Columns(subject auth.Subject, instancePath, engine string) ([]ColumnInfo, error) {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return nil, err
	}
	db, err := uc.series(instancePath, engine)
	if err != nil {
		return nil, err
	}
	keys, err := db.SeriesColumns()
	if err != nil {
		return nil, err
	}
	out := make([]ColumnInfo, 0, len(keys))
	for _, key := range keys {
		bucket, column, ok := splitColumnKey(key)
		if !ok {
			continue
		}
		col, found, err := db.LookupColumn(bucket, column)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		st := col.Stats()
		out = append(out, ColumnInfo{
			Bucket: bucket, Column: column,
			Scheme: st.Scheme, Decimals: st.Decimals,
			Chunks: st.Chunks, Bytes: st.Bytes,
			MinMillis: st.MinMillis, MaxMillis: st.MaxMillis, HasData: st.HasData,
		})
	}
	return out, nil
}

// CountColumn returns the exact number of stored points in a column. This
// performs a full scan (O(points)); the UI caches the result in page state so
// the scan runs only once per page visit.
func (uc UseCases) CountColumn(subject auth.Subject, instancePath, engine, bucket, column string) (int64, error) {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return 0, err
	}
	col, err := uc.column(instancePath, engine, bucket, column)
	if err != nil {
		return 0, err
	}
	return col.Count()
}

// SeriesM4 reads a numeric column, downsampled with M4 to at most 4*Width points,
// over the inclusive [MinMillis, MaxMillis] range. It is constant-memory over the
// raw series regardless of its size.
func (uc UseCases) SeriesM4(subject auth.Subject, req SeriesRequest) ([]SeriesPoint, error) {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return nil, err
	}
	col, err := uc.column(req.Instance, req.Engine, req.Bucket, req.Column)
	if err != nil {
		return nil, err
	}
	if col.Schema().Scheme != tsdb.SchemeDecimal {
		return nil, fmt.Errorf("column %q/%q is not numeric", req.Bucket, req.Column)
	}
	width := req.Width
	if width <= 0 {
		width = 200
	}
	if width > maxChartWidth {
		width = maxChartWidth
	}
	if req.MaxMillis < req.MinMillis {
		return nil, nil
	}

	rng := timeseries.NewRange(timeseries.UnixMilli(req.MinMillis), timeseries.UnixMilli(req.MaxMillis), time.UTC)
	var out []SeriesPoint
	for p := range timeseries.M4(col.IterF64(req.MinMillis, req.MaxMillis), rng, width) {
		out = append(out, SeriesPoint{Millis: int64(p.X), Value: float64(p.Y)})
	}
	return out, nil
}

// StringWindow reads up to Limit (time, value) rows of a string/enum column in
// the inclusive range, in ascending time order.
func (uc UseCases) StringWindow(subject auth.Subject, req StringWindowRequest) ([]StringRow, error) {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return nil, err
	}
	col, err := uc.column(req.Instance, req.Engine, req.Bucket, req.Column)
	if err != nil {
		return nil, err
	}
	scheme := col.Schema().Scheme
	if scheme != tsdb.SchemeString && scheme != tsdb.SchemeEnum {
		return nil, fmt.Errorf("column %q/%q is not a string/enum column", req.Bucket, req.Column)
	}
	limit := req.Limit
	if limit <= 0 || limit > DefaultWindowLimit {
		limit = DefaultWindowLimit
	}

	var out []StringRow
	err = col.ScanString(req.MinMillis, req.MaxMillis, func(ts []int64, vals []string) bool {
		for i := range ts {
			out = append(out, StringRow{Millis: ts[i], Value: vals[i]})
			if len(out) >= limit {
				return false
			}
		}
		return len(out) < limit
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DeletePoint tombstones a single point at ts. Knife tool.
func (uc UseCases) DeletePoint(subject auth.Subject, instancePath, engine, bucket, column string, ts int64) error {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return err
	}
	col, err := uc.column(instancePath, engine, bucket, column)
	if err != nil {
		return err
	}
	return col.Delete(ts)
}

// DeleteSeriesRange tombstones all points in [min,max]. Knife tool.
func (uc UseCases) DeleteSeriesRange(subject auth.Subject, instancePath, engine, bucket, column string, min, max int64) error {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return err
	}
	col, err := uc.column(instancePath, engine, bucket, column)
	if err != nil {
		return err
	}
	return col.DeleteRange(min, max)
}

// FlushColumn forces flush + compaction of a column. Knife tool.
func (uc UseCases) FlushColumn(subject auth.Subject, instancePath, engine, bucket, column string) error {
	if err := subject.Audit(PermNDBInspector); err != nil {
		return err
	}
	col, err := uc.column(instancePath, engine, bucket, column)
	if err != nil {
		return err
	}
	return col.Flush()
}

func splitColumnKey(key string) (bucket, column string, ok bool) {
	i := strings.IndexByte(key, '/')
	if i <= 0 || i >= len(key)-1 {
		return "", "", false
	}
	return key[:i], key[i+1:], true
}
