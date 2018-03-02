package imagesManager

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/go-sql-driver/mysql"
)

var mainDB *sql.DB

//Save copy image struct into SQL Database, return nil if success
func Save(image Image) (int64, error) {
	if image.Location == "" {
		return 0, fmt.Errorf("image location should not be empty")
	}
	if image.ID == 0 {
		return 0, fmt.Errorf("image ID should not be zero")
	}
	if image.FileName == "" {
		return 0, fmt.Errorf("image FileName should not be empty")
	}

	result, err := mainDB.Exec("INSERT INTO (id, filename, location, createdTime, lastmodified) VALUES (?, ?, ?, ?, ?)", image.ID, image.FileName, image.Location, image.CreatedTime, image.LastModifiedTime)
	if err != nil {
		return 0, fmt.Errorf("sql exec failed, %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return id, fmt.Errorf("get insert id failed, %v", err)
	}

	return id, nil

}

func getLocationID(location string) (uint64, error) {
	sql := "insert into " + locationTable + "(" + attrLocation + ")" + " values(?)"
	res, err := SqlExec(util.GetDB(ModuleInfo.ModuleName), sql, location)
	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok { // Now the error number is accessible directly
			if driverErr.Number == ErDupEntry {
				return getExistLocationID(location)
			}
			return 0, err
		}
	}
	id, err := res.LastInsertId()
	return uint64(id), err

}

func getExistLocationID(location string) (uint64, error) {
	sql := "select " + attrID + " from " + locationTable + " where " + attrLocation + "=?"
	rows, err := SqlQuery(util.GetDB(ModuleInfo.ModuleName), sql, location)
	if err != nil {
		return 0, nil
	}
	defer rows.Close()

	var id uint64
	if rows.Next() {
		err = rows.Scan(&id)
	}
	return id, err
}

func getImageRef(args *getImagesArg) (*imageList, error) {

	locations, err := getLocationMap()

	if err != nil {
		return nil, err
	}

	var condition string
	var params []interface{}
	if args.Keyword != "" {
		condition += "where " + attrFileName + " like %?% "
		params = append(params, args.Keyword)
	}
	condition += "order by " + args.Order + " desc "
	condition += "limit " + strconv.FormatInt(args.Limit, 10)
	condition += " offset " + strconv.FormatInt(args.Page, 10)

	sqlString := "select " + attrID + "," + attrFileName + "," + attrSize + "," + attrLocationID + "," + attrCreateTime + "," + attrLatestUpdate + "," + attrAnswerID
	sqlString += " from (select * from " + imageTable + " " + condition + ") as a left join " + relationTable + " as b on " +
		"a." + attrID + "=b." + attrImageID

	rows, err := SqlQuery(util.GetDB(ModuleInfo.ModuleName), sqlString, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	il := &imageList{CurPage: uint64(args.Page)}

	var counter uint64
	var ii *imageInfo
	var lastID uint64
	for rows.Next() {
		var id, locationID uint64
		var size int
		var fileName string
		var createTime, updateTime time.Time
		var answerID sql.NullInt64

		err := rows.Scan(&id, &fileName, &size, &locationID, &createTime, &updateTime, &answerID)
		if err != nil {
			return nil, err
		}
		if lastID != id {
			counter++
			ii = &imageInfo{ImageID: id, Title: fileName, Size: size,
				CreateTime: uint64(createTime.Unix()), LastModified: uint64(updateTime.Unix()),
				URL: locations[locationID],
			}
			il.Images = append(il.Images, ii)
		}
		if answerID.Valid {
			ii.answerID = append(ii.answerID, uint64(answerID.Int64))
		}
	}

	il.Total = counter
	il.CurPage = uint64(args.Page)

	return il, nil
}

func getAnswerRef(answerID []interface{}) ([]*questionInfo, error) {

	tagMap, err := getTagMap()
	if err != nil {
		return nil, err
	}

	qis := make([]*questionInfo, 0)

	if len(answerID) > 0 {
		db := util.GetMainDB()

		// select a.Question_Id,a.Answer_Id,Tag_Id,c.Content,CategoryId from vipshop_answer as a left join vipshop_answertag as b on a.Answer_Id = b.Answer_Id left join vipshop_question as c on a.Question_Id=c.Question_Id where a.Answer_Id in (9733365,9733366,9733367);

		sqlTags := "select a." + attrQID + "," + "a." + attrAnsID + "," + attrTagID + ",c." + attrContent + "," + attrCategoryID +
			" from " + VIPAnswerTable + " as a left join" + VIPAnswerTagTable + " as b on a." + attrAnswerID + "=b." + attrAnswerID +
			" left join " + VIPQuestionTable + " as c on a." + attrQID + "=c." + attrQID
		sqlTags += " where a." + attrAnsID + " in (?" + strings.Repeat(",?", len(answerID)-1) + ")"
		rows, err := SqlQuery(db, sqlTags, answerID...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var lastAnswerID int
		var lastQID int
		var lastQIDCategoryAndQ string
		var qi *questionInfo

		for rows.Next() {
			var qID int
			var answerID int
			var tagID sql.NullInt64
			var delimiter string
			var content string
			var categoryID int

			err := rows.Scan(&qID, &answerID, &tagID, &content, &categoryID)
			if err != nil {
				return nil, err
			}

			if lastAnswerID != answerID || qID != lastQID {
				qi = &questionInfo{QuestionID: qID}
				qis = append(qis, qi)
				if qID == lastQID {
					qi.Info = lastQIDCategoryAndQ
				} else {
					lastQID = qID
					categories, err := GetCategory(categoryID)
					if err != nil {
						return nil, err
					}
					for i := 0; i < 2 && i < len(categories); i++ {
						qi.Info += categories[i] + "/"
					}
					qi.Info += content + "/"
					lastQIDCategoryAndQ = qi.Info
				}
			} else {
				delimiter = ","
			}

			if tagID.Valid {
				qi.Info += delimiter + tagMap[int(tagID.Int64)]
			} else {
				qi.Info += delimiter + "全部"
			}

		}
	}

	return qis, nil
}

func getLocationMap() (map[uint64]string, error) {
	sqlString := "select " + attrID + "," + attrLocation + " from " + locationTable
	rows, err := SqlQuery(util.GetDB(ModuleInfo.ModuleName), sqlString)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locationMap := make(map[uint64]string)

	var id uint64
	var location string
	for rows.Next() {
		err := rows.Scan(&id, &location)
		if err != nil {
			return nil, err
		}
		locationMap[id] = location
	}

	return locationMap, nil
}

func getTagMap() (map[int]string, error) {
	sqlString := "select " + attrTagID + "," + attrTagName + " from " + VIPTagTable
	rows, err := SqlQuery(util.GetMainDB(), sqlString)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tagMap := make(map[int]string)

	var id int
	var name string
	for rows.Next() {
		err := rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}
		tagMap[id] = strings.Trim(name, "#")
	}

	return tagMap, nil

}

func GetCategory(categoryID int) ([]string, error) {
	//maximum level
	const MAXLEVEL = 5
	var categoryName string
	levels := make([]string, MAXLEVEL, MAXLEVEL)
	sqlString := "select " + attrCategoryName + "," + attrParentID + " from " + VIPCategoryTable + " where " + attrCategoryID + "=?"
	for i := 0; i < MAXLEVEL; i++ {
		rows, err := SqlQuery(util.GetMainDB(), sqlString, categoryID)
		if err != nil {
			return nil, err
		}
		rows.Scan(&categoryName, &categoryID)

		if categoryID == 0 {
			levels[i] = categoryName
			break
		}
	}

	levels = levels[:len(levels)]
	reverseSlice(levels)

	return levels, nil
}

func SqlQuery(db *sql.DB, sql string, params ...interface{}) (*sql.Rows, error) {
	if db == nil {
		return nil, errors.New("No module(" + ModuleInfo.ModuleName + ") db connection")
	}
	return db.Query(sql, params...)
}

func GetTx(db *sql.DB) (*sql.Tx, error) {
	if db == nil {
		return nil, errors.New("No module(" + ModuleInfo.ModuleName + ") db connection")
	}
	return db.Begin()
}

func SqlExec(db *sql.DB, sql string, params ...interface{}) (sql.Result, error) {
	if db == nil {
		return nil, errors.New("No module(" + ModuleInfo.ModuleName + ") db connection")
	}

	stmt, err := db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()
	return ExecStmt(stmt, params...)
}

func ExecStmt(stmt *sql.Stmt, params ...interface{}) (sql.Result, error) {
	return stmt.Exec(params...)
}
