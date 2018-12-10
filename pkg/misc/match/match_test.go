package match

import (
	"fmt"
	"io/ioutil"
	"regexp"
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

func benchmarkRegex(b *testing.B) {
	pattern := "帮我.*"
	for n := 0; n < b.N; n++ {
		re := regexp.MustCompile(pattern)
		ret := make([]string, 0, 5)
		for _, sentence := range sentences {
			if re.FindString(sentence) != "" {
				ret = append(ret, sentence)
			}
			if len(ret) >= 5 {
				break
			}
		}
	}
}
func BenchmarkCompare(b *testing.B) {
	b.N = 200000
	b.Run("fuzzy", benchmarkFuzzy)
	b.Run("simple", benchmarkPrefix)
	b.Run("regex", benchmarkRegex)
}
