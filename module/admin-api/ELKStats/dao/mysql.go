package dao

import (
	"database/sql"
	"errors"
	"fmt"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/util"
)

type MySQL struct {
	backendLogDB *sql.DB
	emotibotDB   *sql.DB
}

const (
	TagTypeTable = "tag_type"
	TagsTable    = "tags"
)

func GetTags() (map[string][]data.Tag, error) {
	db := util.GetMainDB()
	if db == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf(`
		SELECT p.code, t.code, t.name
		FROM %s AS p
		INNER JOIN %s AS t
		WHERE t.type = p.id`, TagTypeTable, TagsTable)

	rows, err := db.Query(queryStr)
	if err != nil {
		return nil, err
	}

	tags := make(map[string][]data.Tag)

	for rows.Next() {
		var tagType string
		tag := data.Tag{}
		err = rows.Scan(&tagType, &tag.Code, &tag.Name)
		if err != nil {
			return nil, err
		}

		_, ok := tags[tagType]
		if !ok {
			tags[tagType] = make([]data.Tag, 0)
		}

		tags[tagType] = append(tags[tagType], tag)
	}

	return tags, nil
}
