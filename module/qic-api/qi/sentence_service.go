package qi

import (
	"encoding/hex"
	"encoding/json"
	"time"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
	uuid "github.com/satori/go.uuid"
)

var (
	sentenceDao model.SentenceDao
)

//DataSentence data structures of sentence level
type DataSentence struct {
	ID   uint64     `json:"id"`
	UUID string     `json:"sentence_id,omitempty"`
	Name string     `json:"sentence_name,omitempty"`
	Tags []*DataTag `json:"tags,omitempty"`
}

//DataTag is data struct of tag level
type DataTag struct {
	UUID        string   `json:"tag_id,omitempty"`
	Name        string   `json:"tag_name,omitempty"`
	Type        string   `json:"tag_type,omitempty"`
	PosSentence []string `json:"pos_sentences,omitempty"`
	NegSentence []string `json:"neg_sentences,omitempty"`
}

//GetSentence gets one sentence depends on uuid
func GetSentence(uuid string, enterprise string) (*DataSentence, error) {
	var isDelete int8
	q := &model.SentenceQuery{UUID: []string{uuid}, IsDelete: &isDelete, Enterprise: &enterprise}
	data, err := getSentences(q)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		return data[0], nil
	}
	return nil, nil
}

//GetSentenceList gets list of sentences queried by enterprise
func GetSentenceList(enterprise string, limit int, page int) (uint64, []*DataSentence, error) {
	var isDelete int8
	q := &model.SentenceQuery{Enterprise: &enterprise, Limit: limit, Page: page, IsDelete: &isDelete}
	count, err := sentenceDao.CountSentences(nil, q)
	if err != nil {
		return 0, nil, err
	}

	if count > 0 {
		data, err := getSentences(q)
		return count, data, err
	}

	return count, nil, nil
}

func getSentences(q *model.SentenceQuery) ([]*DataSentence, error) {
	sentences, err := sentenceDao.GetSentences(nil, q)
	if err != nil {
		return nil, err
	}
	numOfSens := len(sentences)
	if numOfSens == 0 {
		return nil, nil
	}
	data := make([]*DataSentence, 0, numOfSens)
	allTagIDs := make([]uint64, 0)
	for i := 0; i < numOfSens; i++ {
		d := &DataSentence{ID: sentences[i].ID, UUID: sentences[i].UUID, Name: sentences[i].Name}
		d.Tags = make([]*DataTag, 0)
		allTagIDs = append(allTagIDs, sentences[i].TagIDs...)
		data = append(data, d)
	}

	//get tags information
	query := model.TagQuery{ID: allTagIDs, Enterprise: &(*q.Enterprise)}
	tags, err := tagDao.Tags(nil, query)
	if err != nil {
		return nil, err
	}
	//transform tag data to map[tag_id] tag
	tagsIDMap := make(map[uint64]*model.Tag)
	for i := 0; i < len(tags); i++ {
		tagsIDMap[tags[i].ID] = &tags[i]
	}

	for i := 0; i < len(data); i++ {
		for _, tagID := range sentences[i].TagIDs {
			if tag, ok := tagsIDMap[tagID]; ok {
				dataTag := &DataTag{UUID: tag.UUID, Name: tag.Name}
				dataTag.PosSentence = make([]string, 0)
				dataTag.NegSentence = make([]string, 0)
				switch tag.Typ {
				case 1:
					dataTag.Type = "keyword"
				case 2:
					dataTag.Type = "dialogue_act"
				case 3:
					dataTag.Type = "user_response"
				default:
					dataTag.Type = "logic"
					logger.Warn.Printf("tag %d has unknow type %d\n", tag.ID, tag.Typ)
				}

				if tag.PositiveSentence != "" {
					err = json.Unmarshal([]byte(tag.PositiveSentence), dataTag.PosSentence)
					if err != nil {
						logger.Error.Printf("umarshal tag positive %s failed. %s", tag.PositiveSentence, err)
						return nil, err
					}
				}
				if tag.NegativeSentence != "" {
					err = json.Unmarshal([]byte(tag.NegativeSentence), dataTag.NegSentence)
					logger.Error.Printf("umarshal tag negative %s failed. %s", tag.NegativeSentence, err)
					return nil, err
				}

				data[i].Tags = append(data[i].Tags, dataTag)
			}
		}
	}

	return data, nil
}

//NewSentence inserts a new sentence
func NewSentence(enterprise string, name string, tagUUID []string) (*DataSentence, error) {

	//query tags ID
	query := model.TagQuery{UUID: tagUUID, Enterprise: &enterprise}
	tags, err := tagDao.Tags(nil, query)
	if err != nil {
		return nil, err
	}

	numOfTags := len(tags)
	if numOfTags != len(tagUUID) {
		logger.Warn.Printf("user input tagUUID [%+v] not equals to tags [%+v] in db\n", tagUUID, tags)
	}

	tagsID := make([]uint64, 0, numOfTags)
	for i := 0; i < numOfTags; i++ {
		tagsID[i] = tags[i].ID
	}

	//insert into the sentence
	now := time.Now().Unix()
	uuid, err := uuid.NewV4()
	if err != nil {
		logger.Error.Printf("generate uuid failed. %s\n", err)
		return nil, err
	}
	uuidStr := hex.EncodeToString(uuid[:])
	s := &model.Sentence{IsDelete: 0, Name: name, Enterprise: enterprise,
		CreateTime: now, UpdateTime: now, TagIDs: tagsID, UUID: uuidStr}
	sentenceDao.InsertSentence(nil, s)

	sentence := &DataSentence{UUID: uuidStr, Name: name}
	return sentence, nil
}
