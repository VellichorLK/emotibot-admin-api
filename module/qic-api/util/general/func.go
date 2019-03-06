package general

import (
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func UUID() (uuidStr string, err error) {
	uuidObj, err := uuid.NewV4()
	if err != nil {
		return
	}

	uuidStr = uuidObj.String()
	uuidStr = strings.Replace(uuidStr, "-", "", -1)
	return
}

func ParseID(r *http.Request) (id string) {
	vars := mux.Vars(r)
	return vars["id"]
}

func IsNil(t interface{}) bool {
	defer func() { recover() }()
	return t == nil || reflect.ValueOf(t).IsNil()
}

func StringsToRunes(ss []string) [][]rune {
	words := make([][]rune, len(ss))
	for idx, s := range ss {
		// ignore empty string
		// ignroe empty string will cause Index out of error in goahocorasick.Machine Build
		if s == "" {
			continue
		}
		word := []rune(s)
		words[idx] = word
	}
	return words
}
