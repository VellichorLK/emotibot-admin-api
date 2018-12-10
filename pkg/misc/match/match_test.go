package match

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

var sentences = []string{}
var matchers = map[MatcherMode]*Matcher{}

func init() {
	content, err := ioutil.ReadFile("./test.input")
	if err != nil {
		panic(err.Error())
	}
	sentences = strings.Split(string(content), "\n")
	matchers[FuzzyMode] = New(sentences, FuzzyMode)
	matchers[PrefixMode] = New(sentences, PrefixMode)
}

func BenchmarkPrefixInit(b *testing.B) {
	for n := 0; n < b.N; n++ {
		New(sentences, PrefixMode)
	}
}

func TestPrefixFindNSentence(t *testing.T) {
	input := []string{"我要怎麼借款", "我要退錢", "我要退錢行嗎", "我要退貨"}
	matcher := New(input, PrefixMode)
	ret := matcher.FindNSentence("我要退", 1)
	fmt.Printf("%+v\n", ret)
}

func benchmarkPrefix(b *testing.B) {
	pattern := "帮我"
	for n := 0; n < b.N; n++ {
		_ = matchers[PrefixMode].FindNSentence(pattern, 5)
	}
}

func benchmarkFuzzy(b *testing.B) {
	pattern := "帮我"
	for n := 0; n < b.N; n++ {
		_ = matchers[FuzzyMode].FindNSentence(pattern, 5)
	}
}

func BenchmarkCompare(b *testing.B) {
	b.N = 100000
	b.Run("fuzzy", benchmarkFuzzy)
	b.Run("simple", benchmarkPrefix)
}
