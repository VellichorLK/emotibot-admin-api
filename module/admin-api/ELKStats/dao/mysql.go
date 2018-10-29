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

func GetTags() (map[string]map[string][]data.Tag, error) {
	db := util.GetMainDB()
	if db == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf(`
		SELECT p.code, t.code, t.name, t.app_id
		FROM %s AS p
		INNER JOIN %s AS t
		WHERE t.type = p.id`, TagTypeTable, TagsTable)

	rows, err := db.Query(queryStr)
	if err != nil {
		return nil, err
	}

	tags := make(map[string]map[string][]data.Tag)

	for rows.Next() {
		var tagType string
		var tagAppID string
		tag := data.Tag{}
		err = rows.Scan(&tagType, &tag.Code, &tag.Name, &tagAppID)
		if err != nil {
			return nil, err
		}

		_, ok := tags[tagAppID]
		if !ok {
			tags[tagAppID] = make(map[string][]data.Tag)
		}

		_, ok = tags[tagAppID][tagType]
		if !ok {
			tags[tagAppID][tagType] = make([]data.Tag, 0)
		}

		tags[tagAppID][tagType] = append(tags[tagAppID][tagType], tag)
	}

	return tags, nil
}
