package data

// use tf-idf for tag extraction

import (
	"sort"
)

type Segment struct {
	text   string
	weight float64
}

func ExtractTags(clusterPos []string, top int) (tags Segments) {
	freqMap := make(map[string]float64)
	total := 0.0

	for _, w := range clusterPos {
		if _, ok := StopWord[w]; ok {
			continue
		}
		if w == "" || w == " " {
			continue
		}
		freqMap[w] += 1.0
		total++
	}

	for k := range freqMap {
		freqMap[k] /= total
	}

	ss := make(Segments, 0)
	for k, v := range freqMap {
		tmp, _ := IDFCache.LoadOrStore(k, 10.0)
		idf := tmp.(float64)
		s := Segment{text: k, weight: idf * v}
		ss = append(ss, s)
	}

	sort.Sort(sort.Reverse(ss))
	if len(ss) > top {
		tags = ss[:top]
	} else {
		tags = ss
	}
	return tags
}

type Segments []Segment

func (s Segment) Text() string {
	return s.text
}

func (s Segment) Weight() float64 {
	return s.weight
}

func (ss Segments) Len() int {
	return len(ss)
}

func (ss Segments) Less(i, j int) bool {
	if ss[i].weight == ss[j].weight {
		return ss[i].text < ss[j].text
	}
	return ss[i].weight < ss[j].weight
}

func (ss Segments) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}
