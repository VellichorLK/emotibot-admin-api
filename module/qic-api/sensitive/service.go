package sensitive

import (
	"github.com/anknown/ahocorasick"
)

var dao sensitiveDao = &sensitiveDAOImpl{}

func IsSensitive(content string) ([]string, error) {
	matched := []string{}
	words, err := dao.GetSensitiveWords()
	if err != nil {
		return matched, err
	}

	rwords := stringsToRunes(words)

	m := new(goahocorasick.Machine)
    if err = m.Build(rwords); err != nil {
        return matched, err
    }

    terms := m.MultiPatternSearch([]rune(content), false)
    for _, t := range terms {
		matched = append(matched, string(t.Word))
	}
	
	return matched, nil
}

func stringsToRunes(ss []string) ([][]rune) {
	words := make([][]rune, len(ss), len(ss))
	for idx, s := range ss {
		word := []rune(s)
		words[idx] = word
	}
	return words
}