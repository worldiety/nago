package ui

import (
	"reflect"
	"testing"
)

func Test_marshalJSON(t *testing.T) {
	type args struct {
		v any
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"hd", args{HorizontalDivider{}}, []byte(`{"type":"HorizontalDivider"}`), false},
		{"l1l", args{ListItem1L{Headline: "abc"}}, []byte(`{"type":"ListItem1L","headline":"abc","leadingIcon":null,"actionEvent":null}`), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshalJSON(tt.args.v)
			t.Log(string(got))
			if (err != nil) != tt.wantErr {
				t.Errorf("marshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("marshalJSON() got = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
