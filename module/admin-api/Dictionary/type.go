package Dictionary

import (
	"fmt"
	"strings"
	"time"
)

const (
	StatusRunning Status = "running"
	StatusFinish  Status = "finish"
	StatusFail    Status = "fail"
)

type Status string

type StatusInfo struct {
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	Message   *string   `json:"message"`
}

// DownloadMeta is the struct describe file which can be downloaded
type DownloadMeta struct {
	UploadTime time.Time `json:"time"`
	UploadFile string    `json:"filename"`
}

type WordBank struct {
	ID   *int   `json:"id,omitempty"`
	Name string `json:"name"`
	// 0: directory, 1:wordbank
	Type         int         `json:"type"`
	Children     []*WordBank `json:"children"`
	SimilarWords string      `json:"similar_words,omitempty"`
	Answer       string      `json:"answer,omitempty"`
}

type WordBankRow struct {
	Level1       string `xlsx:"0"`
	Level2       string `xlsx:"1"`
	Level3       string `xlsx:"2"`
	Level4       string `xlsx:"3"`
	Name         string `xlsx:"4"`
	SimilarWords string `xlsx:"5"`
	Answer       string `xlsx:"6"`
}

func (row WordBankRow) ToString() string {
	similars := strings.Split(strings.Replace(row.SimilarWords, "，", ",", -1), ",")
	trimSimilars := make([]string, len(similars))
	for idx := range trimSimilars {
		trimSimilars[idx] = strings.TrimSpace(similars[idx])
	}
	return fmt.Sprintf("%s>%s>%s>%s\t%s\t%s",
		row.Level1, row.Level2, row.Level3, row.Level4, row.Name,
		strings.Join(trimSimilars, "\t"))
}

func (row WordBankRow) FillLevel(lastRow WordBankRow) WordBankRow {
	newRow := row
	hasLevel := false
	if newRow.Level4 == "" {
		if hasLevel == true {
			newRow.Level4 = lastRow.Level4
		}
	} else {
		hasLevel = true
	}
	if newRow.Level3 == "" {
		if hasLevel == true {
			newRow.Level3 = lastRow.Level3
		}
	} else {
		hasLevel = true
	}
	if newRow.Level2 == "" {
		if hasLevel == true {
			newRow.Level2 = lastRow.Level2
		}
	} else {
		hasLevel = true
	}
	if newRow.Level1 == "" {
		if hasLevel == true {
			newRow.Level1 = lastRow.Level1
		}
	} else {
		hasLevel = true
	}

	if hasLevel == false {
		newRow.Level1 = lastRow.Level1
		newRow.Level2 = lastRow.Level2
		newRow.Level3 = lastRow.Level3
		newRow.Level4 = lastRow.Level4
	}
	return newRow
}

func (row WordBankRow) GetPath() string {
	paths := []string{}
	for true {
		if row.Level1 == "" {
			break
		}
		paths = append(paths, strings.TrimLeft(row.Level1, "*"))

		if row.Level2 == "" {
			break
		}
		paths = append(paths, strings.TrimLeft(row.Level2, "*"))

		if row.Level3 == "" {
			break
		}
		paths = append(paths, strings.TrimLeft(row.Level3, "*"))

		if row.Level4 == "" {
			break
		}
		paths = append(paths, strings.TrimLeft(row.Level4, "*"))
		break
	}

	return strings.Join(paths, "/")
}

func (row WordBankRow) GetAllPathReadOnlyInfo() map[string]bool {
	ret := map[string]bool{}
	paths := []string{}
	checkStr := ""
	for true {
		if row.Level1 == "" {
			break
		}
		checkStr = row.Level1
		paths = append(paths, strings.TrimLeft(checkStr, "*"))
		ret[strings.Join(paths, "/")] = (strings.TrimLeft(checkStr, "*") != checkStr)

		if row.Level2 == "" {
			break
		}
		checkStr = row.Level2
		paths = append(paths, strings.TrimLeft(checkStr, "*"))
		ret[strings.Join(paths, "/")] = (strings.TrimLeft(checkStr, "*") != checkStr)

		if row.Level3 == "" {
			break
		}
		checkStr = row.Level3
		paths = append(paths, strings.TrimLeft(checkStr, "*"))
		ret[strings.Join(paths, "/")] = (strings.TrimLeft(checkStr, "*") != checkStr)

		if row.Level4 == "" {
			break
		}
		checkStr = row.Level4
		paths = append(paths, strings.TrimLeft(checkStr, "*"))
		ret[strings.Join(paths, "/")] = (strings.TrimLeft(checkStr, "*") != checkStr)
		break
	}

	return ret
}

func (row WordBankRow) IsReadOnly() bool {
	checkStr := ""
	for true {
		if row.Level1 == "" {
			break
		}
		checkStr = row.Level1

		if row.Level2 == "" {
			break
		}
		checkStr = row.Level2

		if row.Level3 == "" {
			break
		}
		checkStr = row.Level3

		if row.Level4 == "" {
			break
		}
		checkStr = row.Level4
		break
	}

	return strings.TrimLeft(checkStr, "*") != checkStr
}

func (row WordBankRow) GetParentPath() string {
	paths := []string{}
	for true {
		if row.Level1 == "" {
			break
		}

		if row.Level2 == "" {
			break
		}
		paths = append(paths, strings.TrimLeft(row.Level1, "*"))

		if row.Level3 == "" {
			break
		}
		paths = append(paths, strings.TrimLeft(row.Level2, "*"))

		if row.Level4 == "" {
			break
		}
		paths = append(paths, strings.TrimLeft(row.Level3, "*"))
		break
	}

	return strings.Join(paths, "/")
}

func (row *WordBankRow) ReadFromRow(vals []string) {
	for cellIdx, value := range vals {
		switch cellIdx {
		case 0:
			row.Level1 = value
		case 1:
			row.Level2 = value
		case 2:
			row.Level3 = value
		case 3:
			row.Level4 = value
		case 4:
			row.Name = strings.TrimSpace(value)
		case 5:
			row.SimilarWords = strings.Replace(value, "，", ",", -1)
		case 6:
			row.Answer = value
		}
	}
}

func (row WordBankRow) IsExtendedRow() bool {
	return row.Level1 == "" &&
		row.Level2 == "" &&
		row.Level3 == "" &&
		row.Level4 == "" &&
		row.Name == "" &&
		row.SimilarWords != ""
}

type WordBankClassV3 struct {
	ID           int                `json:"cid"`
	Name         string             `json:"name"`
	Wordbank     []*WordBankV3      `json:"wordbanks"`
	Children     []*WordBankClassV3 `json:"children"`
	Editable     bool               `json:"editable"`
	IntentEngine bool               `json:"ie_flag"`
	RuleEngine   bool               `json:"re_flag"`
}

type WordBankV3 struct {
	ID           int      `json:"wid"`
	Name         string   `json:"name"`
	Editable     bool     `json:"editable"`
	SimilarWords []string `json:"similar_words"`
	Answer       string   `json:"answer"`
}

type WordBankClassRowV3 struct {
	ID           int
	Name         string
	Parent       *int
	Editable     int
	IntentEngine int
	RuleEngine   int
}

type WordBankRowV3 struct {
	ID           int
	Name         string
	ClassID      int
	SimilarWords string
	Answer       string
}
