package scratch

import (
	"reflect"
	"testing"
)

func TestPresentationOverview(t *testing.T) {
	type args struct {
		r PersonRepo
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PresentationOverview(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PresentationOverview() = %v, want %v", got, tt.want)
			}
		})
	}
}
