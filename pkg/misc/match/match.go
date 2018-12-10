package match

import (
	"fmt"

	"github.com/sahilm/fuzzy"
)

type wordNode struct {
	Word       rune
	IsSentence bool
	Children   map[rune]*wordNode
	Parent     *wordNode
	First      *wordNode
	Sibling    *wordNode
}

// MatcherMode decide the method matcher use to match sentences
type MatcherMode int

const (
	// FuzzyMode will use github.com/sahilm/fuzzy lib to do fuzzy search
	FuzzyMode MatcherMode = 0
	// PrefixMode will search for prefix matching
	PrefixMode MatcherMode = 1
)

// Matcher is the return structure of match package
type Matcher struct {
	// root is used only in PrefixMode
	root *wordNode

	sentences   []string
	matcherMode MatcherMode
}

// New will return an new matcher of input sentences,
// matcherMode must be Fuzzy or Prefix
func New(sentences []string, matcherMode MatcherMode) *Matcher {
	matcher := &Matcher{}
	matcher.initPrefixTree(sentences)
	matcher.sentences = sentences
	matcher.matcherMode = matcherMode
	return matcher
}

// Init will setup a the data structure
func (matcher *Matcher) initPrefixTree(sentences []string) {
	matcher.root = &wordNode{
		' ', false, map[rune]*wordNode{}, nil, nil, nil,
	}
	for _, sentence := range sentences {
		current := matcher.root
		for _, r := range sentence {
			if _, ok := current.Children[r]; !ok {
				current.Children[r] = &wordNode{
					r, false, map[rune]*wordNode{}, current, nil, nil,
				}
			}
			current = current.Children[r]
		}
		current.IsSentence = true
	}
	setSibling(matcher.root)
}

func (matcher *Matcher) findNSentenceByFuzzy(pattern string, n int) []string {
	result := fuzzy.Find(pattern, matcher.sentences)

	ret := make([]string, min(result.Len(), n))
	for idx, match := range result {
		if idx >= n {
			break
		}
		ret[idx] = match.Str
	}
	return ret
}

func (matcher *Matcher) findNSentenceByPrefix(pattern string, n int) []string {
	ret := make([]string, 0, 10)
	prefix := []rune{}

	findRoot := matcher.root
	for _, r := range pattern {
		if next, ok := findRoot.Children[r]; ok {
			findRoot = next
			prefix = append(prefix, r)
		} else {
			return ret
		}
	}

	prefixStr := string(prefix)
	current := findRoot
	str := []rune{}
	history := map[*wordNode]bool{}
	for true {
		if current.IsSentence && !history[current] {
			ret = append(ret, prefixStr+string(str))
			if len(ret) >= n {
				break
			}
		}
		orig := current
		if current.First != nil && !history[current] {
			current = current.First
			str = append(str, current.Word)
		} else if current.Sibling != nil {
			current = current.Sibling
			str[len(str)-1] = current.Word
		} else if current.Parent != nil && current != findRoot {
			current = current.Parent
			str[len(str)-1] = '\x00'
			str = str[:len(str)-1]
		} else {
			break
		}
		history[orig] = true
		if current == findRoot {
			break
		}
	}

	return ret
}

// FindNSentence will find at most N sentence start with pattern
func (matcher *Matcher) FindNSentence(pattern string, n int) []string {
	var ret []string
	if matcher.matcherMode == FuzzyMode {
		ret = matcher.findNSentenceByFuzzy(pattern, n)
	} else if matcher.matcherMode == PrefixMode {
		ret = matcher.findNSentenceByPrefix(pattern, n)
	} else {
		ret = []string{}
	}
	return ret
}

func (node wordNode) Print() {
	fmt.Printf("%q, %t\n", node.Word, node.IsSentence)
	fmt.Printf("\tParent: %p, Sibling: %p\n", node.Parent, node.Sibling)
	for key := range node.Children {
		fmt.Printf("\t\t%q:%p\n", key, node.Children[key])
	}
}

func setSibling(node *wordNode) {
	var last *wordNode
	for _, child := range node.Children {
		if last != nil {
			last.Sibling = child
		} else {
			node.First = child
		}
		last = child
	}

	for _, child := range node.Children {
		if child != nil {
			setSibling(child)
		}
	}
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
