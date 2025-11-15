// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package timeseries

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func TestRange_Interval(t *testing.T) {
	tests := []struct {
		name    string
		r       Range
		wantMin UnixMilli
		wantMax UnixMilli
		wantErr bool
	}{
		{
			"UnixMilli",
			"[2020-11-13 12:55:52,2020-11-13 12:56:26]@Etc/UTC",
			1605272152000,
			1605272186000,
			false,
		},

		{
			"UnixMilli+whitespace",
			"  [  2020-11-13 12:55:52  ,   2020-11-13 12:56:26   ]  @  Etc/UTC  ",
			1605272152000,
			1605272186000,
			false,
		},

		{
			"germany-inc-both",
			"[2020-11-13 14:15:00,2020-11-13 14:20:00]@Europe/Berlin",
			1605273300000,
			1605273600000,
			false,
		},

		{
			"germany-exl-left",
			"(2020-11-13 14:15:00,2020-11-13 14:20:00]@Europe/Berlin",
			1605273300001,
			1605273600000,
			false,
		},

		{
			"germany-exl-right",
			"[2020-11-13 14:15:00,2020-11-13 14:20:00)@Europe/Berlin",
			1605273300000,
			1605273599999,
			false,
		},

		{
			"germany-exl-both",
			"(2020-11-13 14:15:00,2020-11-13 14:20:00)@Europe/Berlin",
			1605273300001,
			1605273599999,
			false,
		},

		{
			"err-0",
			"([2020-11-13 14:15:00,2020-11-13 14:20:00)@Europe/Berlin",
			-1,
			-1,
			true,
		},

		{
			"err-1",
			"(2020-11-13 14:15:00,2020-11-13 14:20:00)@",
			-1,
			-1,
			true,
		},

		{
			"err-2",
			"(2020-11-13 14:15:00,,2020-11-13 14:20:00)@",
			-1,
			-1,
			true,
		},

		{
			"empty",
			"",
			math.MinInt64,
			math.MaxInt64,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotMax, err := tt.r.Interval()
			if (err != nil) != tt.wantErr {
				t.Errorf("Interval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMin != tt.wantMin {
				t.Errorf("Interval() gotMin = %v, want %v", gotMin, tt.wantMin)
			}
			if gotMax != tt.wantMax {
				t.Errorf("Interval() gotMax = %v, want %v", gotMax, tt.wantMax)
			}
		})
	}
}

func TestTimeSeries_Range1(t *testing.T) {
	r := Range("[2020-04-01 02:00:00,2020-04-01 07:00:00)@Europe/Berlin")
	min, max, err := r.Interval()
	if err != nil {
		t.Fatal(err)
	}

	berlin := must(time.LoadLocation("Europe/Berlin"))
	a := must(time.ParseInLocation("2006-01-02 15:04:00", "2020-04-01 02:00:00", berlin))
	b := must(time.ParseInLocation("2006-01-02 15:04:00", "2020-04-01 07:00:00", berlin))

	if UnixMilli(a.UnixMilli()) != min {
		t.Fatalf("expected %d but got %d", min, a.Unix())
	}

	if v := UnixMilli(b.UnixMilli()) - 1; max != v {
		t.Fatalf("expected %d but got %d", max, v)
	}

	ts := Series[I64]{
		{X: UnixMilli(a.UnixMilli()), Y: 2},
		{X: UnixMilli(b.UnixMilli()), Y: 3}, // but be kept, is exclusive
	}

	ts.Fuse()

	ts.Delete(min, max)

	if l := len(ts); l != 1 {
		t.Fatalf("expected 0 but got: %d", l)
	}

	if ts[0].X != UnixMilli(b.UnixMilli()) {
		t.Fatalf("expected %d but got %d", b.UnixMilli(), ts[0].X)
	}
}

func TestTimeSeries_Range2(t *testing.T) {

	berlin := must(time.LoadLocation("Europe/Berlin"))
	a := must(time.ParseInLocation("2006-01-02 15:04:00", "2020-04-01 02:00:00", berlin))
	b := must(time.ParseInLocation("2006-01-02 15:04:00", "2020-04-01 07:00:00", berlin))

	type test struct {
		Name     string
		Range    Range
		Expected []time.Time
	}

	table := []test{
		{
			Name:  "both exclusive, so both are kept",
			Range: "(2020-04-01 02:00:00,2020-04-01 07:00:00)@Europe/Berlin",
			Expected: []time.Time{
				a, b,
			},
		},

		{
			Name:     "both inclusive, so both are removed",
			Range:    "[2020-04-01 02:00:00,2020-04-01 07:00:00]@Europe/Berlin",
			Expected: []time.Time{},
		},

		{
			Name:  "min/left is exclusive, so a is kept",
			Range: "(2020-04-01 02:00:00,2020-04-01 07:00:00]@Europe/Berlin",
			Expected: []time.Time{
				a,
			},
		},

		{
			Name:  "max/right is exclusive, so be is kept",
			Range: "[2020-04-01 02:00:00,2020-04-01 07:00:00)@Europe/Berlin",
			Expected: []time.Time{
				b,
			},
		},
	}

	for _, entry := range table {
		min, max, err := entry.Range.Interval()
		if err != nil {
			t.Fatal(err)
		}

		ts := Series[I64]{
			{X: UnixMilli(a.UnixMilli())},
			{X: UnixMilli(b.UnixMilli())},
		}

		ts.Fuse()
		ts.Delete(min, max)

		tsExpect := Series[I64]{}
		for _, v := range entry.Expected {
			tsExpect = append(tsExpect, PI{
				X: UnixMilli(v.UnixMilli()),
			})
		}

		if !reflect.DeepEqual(ts, tsExpect) {
			t.Fatalf("unexpected case %s:\n%v\n%v", entry.Name, ts, tsExpect)
		}
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}
