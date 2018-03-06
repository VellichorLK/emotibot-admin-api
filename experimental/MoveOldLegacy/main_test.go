package main

import (
	"testing"
)

func TestGetFilesInDir(t *testing.T) {
	type testCase struct {
		filters []filter
		expect  int
	}
	testCases := map[string]testCase{
		"NoFilter": testCase{
			[]filter{},
			2,
		},
		"FilterPNG": testCase{
			[]filter{newExtfilter(map[string]bool{
				".png": true,
			})},
			0,
		},
		"FilterJPEG": testCase{
			[]filter{newExtfilter(map[string]bool{
				".jpg":  true,
				".jpeg": true,
			})},
			1,
		},
	}
	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			infos, err := getFilesInDir("./testdata", tt.filters...)
			if err != nil {
				t.Fatal(err)
			}
			if len(infos) != tt.expect {
				t.Errorf("expect %d file in the dir but got %d", tt.expect, len(infos))
			}
		})
	}

}
