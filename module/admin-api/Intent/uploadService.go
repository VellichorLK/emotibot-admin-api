package Intent

import (
	"errors"

	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/tealeg/xlsx"
	"github.com/siongui/gojianfan"
)

func ParseIntentsFromXLSX(file []byte) (ret map[string][]string, err error) {
	xlsxFile, err := xlsx.OpenBinary(file)
	if err != nil {
		return nil, err
	}

	sheets := xlsxFile.Sheets
	if sheets == nil {
		return nil, errors.New(util.Msg["SheetError"])
	}

	if len(sheets) < 1 {
		return nil, errors.New(util.Msg["SheetError"])
	}

	ret = make(map[string][]string)

	for _, sheet := range sheets {
		sentences := make([]string, 0)

		for _, row := range sheet.Rows {
			cells := row.Cells
			if cells != nil && len(cells) > 0 {
				sentence := cells[0].String()
				if sentence != "" {
					sentences = append(sentences, gojianfan.T2S(sentence))
				}
			}
		}

		ret[gojianfan.T2S(sheet.Name)] = sentences
    }

    return
}
