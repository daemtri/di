package jsonconfig

import (
	"flag"
	"testing"

	"github.com/tidwall/gjson"
)

func Test_parseJSON(t *testing.T) {
	type args struct {
		fs     *flag.FlagSet
		prefix string
		result *gjson.Result
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parseJSON(tt.args.fs, tt.args.prefix, tt.args.result)
		})
	}
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name     string
		fs       *flag.FlagSet
		json     string
		wantErr  bool
		wantFlag map[string]string
	}{
		{
			name: "simple",
			fs: func() *flag.FlagSet {
				fs := flag.NewFlagSet("simple", flag.ContinueOnError)
				fs.String("a", "b", "test a")
				return fs
			}(),
			wantErr: false,
			json:    `{"a"="c"}`,
			wantFlag: map[string]string{
				"a": "c",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseJSON(tt.fs, tt.json); (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.fs.Parse(nil)
			for key, value := range tt.wantFlag {
				f := tt.fs.Lookup(key)
				if f == nil {
					t.Errorf("ParseJSON() error = f is nil ")
				} else if f.Value.String() != value {
					t.Errorf("ParseJSON() error = value not equal: %s=%s", f.Value, value)
				}
			}
		})
	}
}
