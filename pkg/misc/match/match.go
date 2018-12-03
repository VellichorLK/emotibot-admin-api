package match

import "fmt"

type wordNode struct {
	Word       rune
	IsSentence bool
	Children   map[rune]*wordNode
	Parent     *wordNode
	First      *wordNode
	Sibling    *wordNode
}

// Matcher is the return structure of match package
type Matcher struct {
	root *wordNode
}

// New will return an new matcher of input sentences
func New(sentences []string) *Matcher {
	matcher := &Matcher{}
	matcher.init(sentences)
	return matcher
}

// Init will setup a the data structure
func (matcher *Matcher) init(sentences []string) {
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

// FindNSentence will find at most N sentence start with pattern
func (matcher *Matcher) FindNSentence(pattern string, n int) []string {
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
