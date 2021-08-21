package main

import (
	"reflect"
	"testing"
)

func Test_initTfmBins(t *testing.T) {
	var (
		p_empty   = ""
		p_invalid = "path/to/void"
		p_onepath = "testdata/case1/bin"
		p_allpath = "testdata/case1/bin:testdata/case1/opt/bin:testdata/case1/usr/local/bin"
	)
	tests := []struct {
		name   string
		ospath *string
		want   map[string]string
	}{
		// TODO: Add test cases.
		{
			name:   "Empty path",
			ospath: &p_empty,
			want:   map[string]string{},
		},
		{
			name:   "Invalid path",
			ospath: &p_invalid,
			want:   map[string]string{},
		},
		{
			name:   "One path",
			ospath: &p_onepath,
			want:   map[string]string{"0.11.2": "testdata/case1/bin/terraform-0.11.2"},
		},
		{
			name:   "All paths",
			ospath: &p_allpath,
			want: map[string]string{
				"0.11.1": "testdata/case1/usr/local/bin/terraform-0.11.1",
				"0.11.2": "testdata/case1/bin/terraform-0.11.2",
				"0.12.2": "testdata/case1/usr/local/bin/terraform-0.12.2",
				"0.13.1": "testdata/case1/usr/local/bin/terraform-0.13.1",
				"0.13.6": "testdata/case1/usr/local/bin/terraform-0.13.6",
				"1.0.3":  "testdata/case1/opt/bin/terraform-1.0.3",
				"1.0.4":  "testdata/case1/opt/bin/terraform-1.0.4",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitTfmBins(tt.ospath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initTfmBins() = %v, want %v", got, tt.want)
			}
		})
	}
}
