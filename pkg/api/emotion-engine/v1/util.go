package emotionengine

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// CSVToEmotion is a convention function helper to read a example csv file to Emotion struct.
// csv format header(optional):
//		sentence, tag
// Column 1: training text
// Column 2: tag, to tell which
func CSVToEmotion(csvFile, emotionName string) (Emotion, error) {
	cf, err := os.Open(csvFile)
	if err != nil {
		return Emotion{}, fmt.Errorf("OS Open file failed, %v", err)
	}
	defer cf.Close()
	reader := csv.NewReader(cf)
	var e = Emotion{
		Name:             emotionName,
		PositiveSentence: make([]string, 0, 0),
		NegativeSentence: make([]string, 0, 0),
	}
	var i int
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		i++
		if err != nil {
			return Emotion{}, fmt.Errorf("csv read error: %v", err)
		}

		if len(row) < 2 {
			return Emotion{}, fmt.Errorf("parsing error: row %d is empty", i)
		}
		//skip header line
		if row[0] == "sentence" {
			continue
		}

		if row[1] == "其他" {
			e.NegativeSentence = append(e.NegativeSentence, row[0])
		} else {
			e.PositiveSentence = append(e.PositiveSentence, row[0])
		}
	}
	return e, nil
}
