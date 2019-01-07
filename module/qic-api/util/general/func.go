package general

import (
	"github.com/satori/go.uuid"
	"math/rand"
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
