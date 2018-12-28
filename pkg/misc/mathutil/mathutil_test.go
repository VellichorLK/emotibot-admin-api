package mathutil

import "testing"

func TestMaxInt(t *testing.T) {
	type testcase struct {
		input  []int
		output int
	}
	testsTable := map[string]testcase{
		"positive int": testcase{
			input:  []int{1, 2},
			output: 2,
		},
		"equal": testcase{
			input:  []int{10, 10},
			output: 10,
		},
		"negative int": testcase{
			input:  []int{-10, -1},
			output: -1,
		},
	}

	for name, tc := range testsTable {
		t.Run(name, func(tt *testing.T) {
			if value := MaxInt(tc.input[0], tc.input[1]); tc.output != value {
				tt.Error("expect output ", tc.output, " but got ", value)
			}
		})
	}
}

func TestMinInt(t *testing.T) {
	type testcase struct {
		input  []int
		output int
	}
	testsTable := map[string]testcase{
		"positive int": testcase{
			input:  []int{2, 1},
			output: 1,
		},
		"equal": testcase{
			input:  []int{10, 10},
			output: 10,
		},
		"negative int": testcase{
			input:  []int{-10, -1},
			output: -10,
		},
	}

	for name, tc := range testsTable {
		t.Run(name, func(tt *testing.T) {
			if value := MinInt(tc.input[0], tc.input[1]); tc.output != value {
				tt.Error("expect output ", tc.output, " but got ", value)
			}
		})
	}
}
