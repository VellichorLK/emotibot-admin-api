package model

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
)

var isIntegration bool

func TestMain(m *testing.M) {
	flag.BoolVar(&isIntegration, "integration", false, "flag for running integration test")
	flag.Parse()
	// a old test compatible
	testTags = seedTags()
	os.Exit(m.Run())
}

func skipIntergartion(t *testing.T) {
	if !isIntegration {
		t.Skip("skip intergration test, please specify -intergation flag.")
	}
	return
}

func newIntegrationTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1)/QISYS?parseTime=true&loc=Asia%2FTaipei")
	if err != nil {
		t.Fatal("expect db open success but got error: ", err)
	}
	return db
}

// Binding take g as the binding subject, and bind the slice the data string into it.
// g must be the struct and is addressable(pointer), data's order also need to be aligned with the order of the subject struct.
// This function is only intended for testing purpose. all error will be panic.
func Binding(g interface{}, data []string) {
	v := reflect.ValueOf(g)
	s := v.Elem()
	for j := 0; j < s.NumField(); j++ {
		f := s.Field(j)
		fieldName := s.Type().Field(j).Name
		switch f.Kind() {
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint8:
			uv, err := strconv.ParseUint(data[j], 10, 64)
			if err != nil {
				panic(fmt.Sprintf("column %d: '%s' parse as %s uint failed, %s", j, data[j], fieldName, err))
			}
			f.SetUint(uv)
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			iv, err := strconv.ParseInt(data[j], 10, 64)
			if err != nil {
				panic(fmt.Sprintf("column %d: '%s' parse as %s int failed, %v", j, data[j], fieldName, err))
			}
			f.SetInt(iv)
		case reflect.Float32, reflect.Float64:
			fv, err := strconv.ParseFloat(data[j], 64)
			if err != nil {
				panic(fmt.Sprintf("column %d: '%s' parse as %s float failed, %s", j, data[j], fieldName, err))
			}
			f.SetFloat(fv)
		case reflect.String:
			f.SetString(data[j])
		case reflect.Bool:
			bv, err := strconv.ParseInt(data[j], 10, 8)
			if err != nil {
				panic(fmt.Sprintf("column %d: '%s' parse as %s int8 failed, %v", j, data[j], fieldName, err))
			}
			if bv == 0 {
				f.SetBool(false)
			} else {
				f.SetBool(true)
			}
		}

	}
}
