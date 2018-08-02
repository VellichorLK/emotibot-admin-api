package intentengine

import (
	"errors"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/localemsg"
	"github.com/siongui/gojianfan"
	"github.com/tealeg/xlsx"
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

	for idx := range sheets {
		if sheets[idx].Name == localemsg.Get("zh-cn", "IntentBF2Sheet1Name") ||
			sheets[idx].Name == localemsg.Get("zh-tw", "IntentBF2Sheet1Name") {
			return parseBF2IntentsFormat(sheets[idx])
		}
	}
	return parseBFOPIntentsFormat(sheets)
}

func parseBF2IntentsFormat(sheet *xlsx.Sheet) (ret map[string][]string, err error) {
	ret = make(map[string][]string)

	rows := sheet.Rows
	if len(rows) <= 1 {
		return ret, nil
	}

	headerRow := rows[0]
	if len(headerRow.Cells) != 2 {
		return nil, errors.New("Invalid header")
	}
	nameIdx := -1
	sentenceIdx := -1
	nameHeader := localemsg.Get("zh-cn", "IntentName")
	setenceHeader := localemsg.Get("zh-cn", "IntentSentence")
	for idx := range headerRow.Cells {
		switch headerRow.Cells[idx].String() {
		case nameHeader:
			nameIdx = idx
		case setenceHeader:
			sentenceIdx = idx
		}
	}
	if nameIdx == -1 || sentenceIdx == -1 {
		return nil, errors.New("Invalid header")
	}

	for _, row := range sheet.Rows {
		cells := row.Cells
		if cells != nil && len(cells) > 1 {
			intent := strings.TrimSpace(cells[nameIdx].String())
			sentence := strings.TrimSpace(cells[sentenceIdx].String())
			if intent != "" && sentence != "" {
				if _, ok := ret[gojianfan.T2S(sheet.Name)]; !ok {
					ret[gojianfan.T2S(sheet.Name)] = []string{}
				}
				ret[gojianfan.T2S(sheet.Name)] = append(
					ret[gojianfan.T2S(sheet.Name)], gojianfan.T2S(sentence))
			}
		}
	}

	return
}

func parseBFOPIntentsFormat(sheets []*xlsx.Sheet) (ret map[string][]string, err error) {
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
