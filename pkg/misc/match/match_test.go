package match

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/sahilm/fuzzy"
)

var sentences = []string{}
var matcher *Matcher

func init() {
	content, err := ioutil.ReadFile("./test.input")
	if err != nil {
		panic(err.Error())
	}
	sentences = strings.Split(string(content), "\n")
	matcher = New(sentences)
}

func benchmarkFuzzy(b *testing.B) {
	pattern := "帮我"
	for n := 0; n < b.N; n++ {
		_ = fuzzy.Find(pattern, sentences)
	}
}

func BenchmarkInit(b *testing.B) {
	for n := 0; n < b.N; n++ {
		New(sentences)
	}
}

func TestFindNSentence(t *testing.T) {
	input := []string{"我要怎麼借款", "我要退錢", "我要退錢行嗎", "我要退貨"}
	matcher := New(input)
	ret := matcher.FindNSentence("我要退", 1)
	fmt.Printf("%+v\n", ret)
}

func benchmarkSimple(b *testing.B) {
	pattern := "帮我"
	for n := 0; n < b.N; n++ {
		_ = matcher.FindNSentence(pattern, 5)
	}
}

func BenchmarkCompare(b *testing.B) {
	b.N = 100000
	b.Run("fuzzy", benchmarkFuzzy)
	b.Run("simple", benchmarkSimple)
}
