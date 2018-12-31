package qi

import (
	"fmt"

	"emotibot.com/emotigo/module/qic-api/util/general"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

var tagTypeDict = map[int8]string{
	0: "default",
	1: "keyword",
	2: "dialogue_act",
	3: "user_response",
}

//Tags is the service for getting the tags json response.
func Tags(entID string, limit, page int) (resp *TagResponse, err error) {
	enterprise := entID
	query := model.TagQuery{
		Enterprise: &enterprise,
		Paging: &model.Pagination{
			Limit: limit,
			Page:  page,
		},
	}
	result, err := tagDao.Tags(nil, query)
	if err != nil {
		return nil, fmt.Errorf("get tag from dao failed, %v", err)
	}
	counts, err := tagDao.CountTags(nil, query)
	if err != nil {
		return nil, fmt.Errorf("get tag count from dao failed, %v", err)
	}
	var tags = make([]tag, 0, len(result))
	for _, t := range result {
		typ, found := tagTypeDict[t.Typ]
		if !found {
			typ = "default"
		}

		tags = append(tags, tag{
			TagID:        t.ID,
			TagName:      t.Name,
			TagType:      typ,
			PosSentences: []byte(t.PositiveSentence),
			NegSentences: []byte(t.PositiveSentence),
		})
	}
	resp = &TagResponse{
		Paging: general.Paging{
			Total: int64(counts),
			Limit: limit,
			Page:  page,
		},
		Data: tags,
	}
	return
}
