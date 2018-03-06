package imagesManager

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
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
	res, err := SqlExec(db, sql, location)
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
	rows, err := SqlQuery(db, sql, location)
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
		condition += "where " + attrFileName + " like ? "
		params = append(params, "%"+args.Keyword+"%")
	}
	condition += "order by " + args.Order + " desc "
	condition += "limit " + strconv.FormatInt(args.Limit, 10)
	condition += " offset " + strconv.FormatInt(args.Page, 10)

	sqlString := "select " + attrID + "," + attrFileName + "," + attrSize + "," + attrLocationID + "," + attrCreateTime + "," + attrLatestUpdate + "," + attrAnswerID
	sqlString += " from (select * from " + imageTable + " " + condition + ") as a left join " + relationTable + " as b on " +
		"a." + attrID + "=b." + attrImageID

	rows, err := SqlQuery(db, sqlString, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	il := &imageList{CurPage: uint64(args.Page)}
	il.Images = make([]*imageInfo, 0)

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
			lastID = id
			ii = &imageInfo{ImageID: id, Title: fileName, Size: size,
				CreateTime: uint64(createTime.Unix()), LastModified: uint64(updateTime.Unix()),
				URL: locations[locationID],
			}
			il.Images = append(il.Images, ii)
		}
		if answerID.Valid {
			ii.answerID = append(ii.answerID, int(answerID.Int64))
			il.answerIDs = append(il.answerIDs, answerID.Int64)
		}
	}

	il.Total = counter
	il.CurPage = uint64(args.Page)

	return il, nil
}

func getAnswerRef(answerID []interface{}, tagMap map[int]string, categoriesMap map[int]*Category) (map[int]*questionInfo, error) {

	qis := make(map[int]*questionInfo, 0)

	if len(answerID) > 0 {

		mainDB := util.GetMainDB()

		// select a.Question_Id,a.Answer_Id,Tag_Id,c.Content,CategoryId from vipshop_answer as a left join vipshop_answertag as b on a.Answer_Id = b.Answer_Id left join vipshop_question as c on a.Question_Id=c.Question_Id where a.Answer_Id in (9733365,9733366,9733367);

		sqlTags := "select a." + attrQID + "," + "a." + attrAnsID + "," + attrTagID + ",c." + attrContent + "," + attrCategoryID +
			" from " + VIPAnswerTable + " as a left join " + VIPAnswerTagTable + " as b on a." + attrAnswerID + "=b." + attrAnswerID +
			" left join " + VIPQuestionTable + " as c on a." + attrQID + "=c." + attrQID
		sqlTags += " where a." + attrAnsID + " in (?" + strings.Repeat(",?", len(answerID)-1) + ")"
		rows, err := SqlQuery(mainDB, sqlTags, answerID...)
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
				lastAnswerID = answerID
				qi = &questionInfo{QuestionID: qID}
				//qis = append(qis, qi)
				qis[answerID] = qi
				if qID == lastQID {
					qi.Info = lastQIDCategoryAndQ
				} else {
					lastQID = qID
					categories, err := GetFullCategory(categoriesMap, categoryID)
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
	rows, err := SqlQuery(db, sqlString)
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

func GetCategories() (map[int]*Category, error) {

	sqlString := "select " + attrCategoryID + "," + attrCategoryName + "," + attrParentID + " from " + VIPCategoryTable

	rows, err := SqlQuery(util.GetMainDB(), sqlString)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make(map[int]*Category)

	var categoryID, parentID int
	var categoryName string
	for rows.Next() {
		err := rows.Scan(&categoryID, &categoryName, &parentID)
		if err != nil {
			return nil, err
		}
		category := &Category{Name: categoryName, ParentID: parentID}
		categories[categoryID] = category
	}

	return categories, nil
}

func getFileNameByImageID(imageIDs []interface{}) ([]string, error) {
	sqlString := "select " + attrFileName + " from " + imageTable + " where " + attrID + " in (?" + strings.Repeat(",?", len(imageIDs)-1) + ")"

	rows, err := SqlQuery(db, sqlString, imageIDs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fileNameList := make([]string, 0)
	var fileName string
	for rows.Next() {
		err := rows.Scan(&fileName)
		if err != nil {
			return nil, err
		}
		fileNameList = append(fileNameList, fileName)
	}
	return fileNameList, nil
}

func createBackupFolder(n int, path string) (string, error) {
	folderName := GetUniqueString(n)
	err := os.Mkdir(path+"/"+folderName, 0755)
	if err != nil {
		return "", err
	}
	return folderName, nil
}

func copyFiles(from string, to string, fileName []string) (int, error) {
	var count int
	for i := 0; i < len(fileName); i++ {

		src := from + "/" + fileName[i]
		dst := to + "/" + fileName[i]

		if _, err := os.Stat(dst); os.IsNotExist(err) {
			err := copy(src, dst)
			if err != nil {
				//if the file doesn't exist, assusme it is deleted somehow.
				if e, ok := err.(*os.PathError); ok && os.IsNotExist(e) {
					util.LogWarn.Printf("opening file %s failed. File not exist.\n", from+"/"+fileName[i])
					continue
				}
				util.LogError.Printf("opening file %s failed. %s\n", fileName[i], err.Error())
				return count, err
			}
			count++
		} else {
			fmt.Println(err)
		}
	}

	return count, nil
}

func copy(from string, to string) error {
	srcFile, err := os.Open(from)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(to, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}

func deleteFiles(prefix string, fileName []string) (int64, error) {
	var count int64

	if prefix != "" {
		prefix += "/"
	}

	for _, name := range fileName {
		err := os.RemoveAll(prefix + name)
		if err != nil {
			//if the file doesn't exist, assusme it is deleted by another goroutine as the same time.
			if e, ok := err.(*os.PathError); ok && os.IsNotExist(e) {
				util.LogWarn.Printf("remove file %s failed. File not exist.\n", name)
				continue
			}
			util.LogError.Printf("remove file %s failed! %s", name, err.Error())
			return count, errors.New("remove file " + name + " failed." + err.Error())
		}
		count++
	}
	return count, nil
}

//return value: insert row id, rows affected, insert file name, error
//this function tries to insert the assigned filename. If it's duplicate, try to make new one and insert
func inputUniqueFileName(tx *sql.Tx, sql string, name string, params []interface{}) (int64, int64, string, error) {
	var id int64
	var counter int
	var affected int64

	ext := path.Ext(name)
	baseName := name[:len(name)-len(ext)]
	fileName := name

	stmt, err := tx.Prepare(sql)
	if err != nil {
		return 0, 0, "", err
	}
	defer stmt.Close()

	for {
		res, err := stmt.Exec(params...)
		if err != nil {
			if driverErr, ok := err.(*mysql.MySQLError); ok {
				if driverErr.Number != ErDupEntry {
					return 0, 0, "", err
				}
			} else {
				return 0, 0, "", err
			}
		} else {
			id, err = res.LastInsertId()
			affected, err = res.RowsAffected()
			break
		}
		counter++
		fileName = baseName + "(" + strconv.Itoa(counter) + ")" + ext
		params[0] = fileName
	}

	return id, affected, fileName, nil
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
