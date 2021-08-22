package main

import (
	"os"
	"reflect"
	"testing"
)

func TestInitVars(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
	}{
		{
			name: "test with env variables",
			envs: map[string]string{
				"PATH":               "testdata/case1/usr/local/bin",
				"DEBUG":              "true",
				"DEFAULT_CONSTRAINT": "= 0.12.1",
			},
		},
		{
			name: "empty env variables",
			envs: map[string]string{
				"PATH":               "",
				"DEBUG":              "",
				"DEFAULT_CONSTRAINT": "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envs {
				if v != "" {
					os.Setenv(k, v)
				} else {
					os.Unsetenv(k)
				}
			}
			InitVars()
			if FlagDebug {
				if tt.envs["DEBUG"] != "true" {
					t.Errorf("InitVars() = debug:%v, want %v", true, false)
				}
			} else {
				if tt.envs["DEBUG"] == "true" {
					t.Errorf("InitVars() = debug:%v, want %v", false, true)
				}
			}
			for k, v := range tt.envs {
				switch k {
				case "PATH":
					if OsPath != v {
						t.Errorf("InitVars() = OsPath:%v, want %v", OsPath, v)
					}
				case "DEFAULT_CONSTRAINT":
					if v == "" {
						if DefaultConstraint != VerConstraint {
							t.Errorf("InitVars(unset) = DefaultConstraint:%v, want %v", DefaultConstraint, VerConstraint)
						}
					} else if DefaultConstraint != v {
						t.Errorf("InitVars(set) = DefaultConstraint:%v, want %v", DefaultConstraint, v)
					}
				}
				if v != "" {
					os.Setenv(k, v)
				} else {
					os.Unsetenv(k)
				}
			}
		})
	}
}

func TestInitTfmBins(t *testing.T) {
	var (
		pEmpty   = ""
		pInvalid = "path/to/void"
		pOnepath = "testdata/case1/bin"
		pAllpath = "testdata/case1/bin:testdata/case1/opt/bin:testdata/case1/usr/local/bin"
	)
	tests := []struct {
		name   string
		ospath *string
		want   map[string]string
	}{
		{
			name:   "Empty path",
			ospath: &pEmpty,
			want:   map[string]string{},
		},
		{
			name:   "Invalid path",
			ospath: &pInvalid,
			want:   map[string]string{},
		},
		{
			name:   "One path",
			ospath: &pOnepath,
			want:   map[string]string{"0.11.2": "testdata/case1/bin/terraform-0.11.2"},
		},
		{
			name:   "All paths",
			ospath: &pAllpath,
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

func TestSelectTfmBin(t *testing.T) {
	type args struct {
		c string
		b map[string]string
	}
	bins := map[string]string{
		"0.0.0": "path/to/terraform-0.0.0",
		"0.1.0": "path/to/terraform-0.1.0",
		"0.1.1": "path/to/terraform-0.1.1",
		"0.1.2": "path/to/terraform-0.1.2",
		"1.1.0": "path/to/terraform-1.1.0",
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1_no_operator_version",
			args: args{
				c: "0.0.0",
				b: bins,
			},
			want: "path/to/terraform-0.0.0",
		},
		{
			name: "test2_exact_version_number",
			args: args{
				c: "= 0.1.1",
				b: bins,
			},
			want: "path/to/terraform-0.1.1",
		},
		{
			name: "test3_rightmost_version",
			args: args{
				c: "~> 0.1.1",
				b: bins,
			},
			want: "path/to/terraform-0.1.2",
		},
		{
			name: "test4_empty_version",
			args: args{
				c: "",
				b: bins,
			},
			want: "path/to/terraform-1.1.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectTfmBin(tt.args.c, tt.args.b); got != tt.want {
				t.Errorf("SelectTfmBin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTfmConfigs(t *testing.T) {
	type args struct {
		workdirDir string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "app1 with version attribute",
			args:    args{workdirDir: "testdata/case2/opt/src/app1"},
			want:    "~> 0.13.1",
			wantErr: false,
		},
		{
			name:    "app2 without version attribute",
			args:    args{workdirDir: "testdata/case2/opt/src/app2"},
			want:    "= 0.13.6",
			wantErr: false,
		},
		{
			name:    "app3 without *.tf files",
			args:    args{workdirDir: "testdata/case2/opt/src/app3"},
			want:    "= 0.13.6",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTfmConfigs(tt.args.workdirDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTfmConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseTfmConfigs() = %v, want %v", got, tt.want)
			}
		})
	}
}
