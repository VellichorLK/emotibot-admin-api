package imagesManager

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/go-sql-driver/mysql"
)

var mainDB *sql.DB

func getLocationID(location string) (uint64, error) {
	sql := "insert into " + locationTable + "(" + attrLocation + ")" + " values(?)"
	res, err := SqlExec(db, sql, location)
	if err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok { // Now the error number is accessible directly
			if driverErr.Number == ErDupEntry {
				return getExistLocationID(location)
			}
		}
		return 0, err
	}
	id, err := res.LastInsertId()
	return uint64(id), err

}

func getExistLocationID(location string) (uint64, error) {
	lastOperation := time.Now()
	sql := "select " + attrID + " from " + locationTable + " where " + attrLocation + "=?"
	rows, err := SqlQuery(db, sql, location)
	if err != nil {
		return 0, nil
	}
	defer rows.Close()
	util.LogInfo.Printf("query image location in getExistLocationID took: %s\n", time.Since(lastOperation))

	var id uint64
	if rows.Next() {
		err = rows.Scan(&id)
	}
	return id, err
}

func getImageRef(args *getImagesArg) (*imageList, error) {
	lastOperation := time.Now()
	locations, err := getLocationMap()

	if err != nil {
		return nil, err
	}
	util.LogInfo.Printf("query image location map in getImageRef took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()


	var condition string
	var params []interface{}
	var totalCount uint64
	if args.Keyword != "" {
		condition += "where " + attrFileName + " like ? "
		params = append(params, "%"+util.EscapeQuery(args.Keyword)+"%")
	}

	countSQL := "select count(*) from " + imageTable + " " + condition
	rows, err := db.Query(countSQL, params...)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		err := rows.Scan(&totalCount)
		if err != nil {
			return nil, err
		}
	}
	rows.Close()
	util.LogInfo.Printf("count images in getImageRef took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	condition += "order by " + args.Order + " desc "
	condition += "limit " + strconv.FormatInt(args.Limit, 10)
	condition += " offset " + strconv.FormatInt(args.Page*args.Limit, 10)

	sqlString := "select a." + attrID + ",a." + attrFileName + ",a." + attrSize + ",a." + attrLocationID + ",a." + attrCreateTime + ",a." + attrLatestUpdate + ",a." + attrRawFileName + ",b." + attrAnswerID
	sqlString += " from (select * from " + imageTable + " " + condition + ") as a left join " + relationTable + " as b on " +
		"a." + attrID + "=b." + attrImageID

	rows, err = SqlQuery(db, sqlString, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	util.LogInfo.Printf("get images in getImageRef took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

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
		var rawFileName sql.NullString
		err := rows.Scan(&id, &fileName, &size, &locationID, &createTime, &updateTime, &rawFileName, &answerID)
		if err != nil {
			return nil, err
		}
		if lastID != id {
			counter++
			lastID = id
			if rawFileName.Valid {

				imageURL, err := url.Parse(locations[locationID] + "/" + url.PathEscape(rawFileName.String))
				if err != nil {
					util.LogWarn.Println(err)
					continue
				}

				ii = &imageInfo{
					ImageID: id, Title: fileName, Size: size,
					CreateTime: uint64(createTime.Unix()), LastModified: uint64(updateTime.Unix()),
					URL: imageURL.String(),
				}
				il.Images = append(il.Images, ii)
			}
		}
		if answerID.Valid {
			ii.answerID = append(ii.answerID, int(answerID.Int64))
			il.answerIDs = append(il.answerIDs, answerID.Int64)
		}
	}

	il.Total = totalCount
	il.CurPage = uint64(args.Page)
	util.LogInfo.Printf("create image structs in getImageRef took: %s\n", time.Since(lastOperation))

	return il, nil
}

func getAnswerRef(answerID []interface{}, tagMap map[int]string, categoriesMap map[int]*Category) (map[int]*questionInfo, error) {

	qis := make(map[int]*questionInfo, 0)
	lastOperation := time.Now()
	if len(answerID) > 0 {

		mainDB := util.GetMainDB()

		// select a.Question_Id,a.Answer_Id,Tag_Id,c.Content,CategoryId from vipshop_answer as a
		// left join vipshop_answertag as b on a.Answer_Id   = b.Answer_Id
		// left join vipshop_question  as c on a.Question_Id = c.Question_Id
		// where a.Answer_Id in (9733365,9733366,9733367);

		sqlTags := "select a." + attrQID + "," + "a." + attrAnsID + "," + attrTagID + ",c." + attrContent + "," + attrCategoryID +
			" from " + VIPAnswerTable + " as a left join " + VIPAnswerTagTable + " as b on a." + attrAnswerID + "=b." + attrAnswerID +
			" left join " + VIPQuestionTable + " as c on a." + attrQID + "=c." + attrQID
		sqlTags += " where a." + attrAnsID + " in (?" + strings.Repeat(",?", len(answerID)-1) + ")"
		rows, err := SqlQuery(mainDB, sqlTags, answerID...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		util.LogInfo.Printf("get images answer reference in getAnswerRef took: %s\n", time.Since(lastOperation))
		lastOperation = time.Now()

		var lastAnswerID int
		//Speedup same question ID's answer
		var lastQID int
		var lastQIDCategoryAndQ string

		var qi *questionInfo

		for rows.Next() {
			var qID int
			var answerID int
			var tagID sql.NullInt64
			var delimiter string
			var content sql.NullString
			var categoryID sql.NullInt64

			err := rows.Scan(&qID, &answerID, &tagID, &content, &categoryID)
			if err != nil {
				return nil, err
			}

			if !content.Valid || !categoryID.Valid {
				util.LogError.Printf("Error: cannot find answer id %d 's question!\n", answerID)
				continue
			}

			if lastAnswerID != answerID || qID != lastQID {
				lastAnswerID = answerID
				qi = &questionInfo{QuestionID: qID}
				if qID == lastQID {
					qi.Info = lastQIDCategoryAndQ
				} else {
					categories, err := GetFullCategory(categoriesMap, int(categoryID.Int64))
					if err != nil {
						util.LogError.Printf("Category %d has error %v\n", categoryID.Int64, err)
						continue
					}
					lastQID = qID
					for i := 0; i < 2 && i < len(categories); i++ {
						qi.Info += categories[i] + "/"
					}
					qi.Info += content.String + "/"
					lastQIDCategoryAndQ = qi.Info
				}
				qis[answerID] = qi
			} else {
				delimiter = ","
			}

			if tagID.Valid {
				qi.Info += delimiter + tagMap[int(tagID.Int64)]
			} else {
				qi.Info += delimiter + "所有维度"
			}

		}
		util.LogInfo.Printf("cerate images answer reference struct in getAnswerRef took: %s\n", time.Since(lastOperation))
	}

	return qis, nil
}

func getLocationMap() (map[uint64]string, error) {
	lastOperation := time.Now()
	sqlString := "select " + attrID + "," + attrLocation + " from " + locationTable
	rows, err := SqlQuery(db, sqlString)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	util.LogInfo.Printf("get location in getLocationMap took: %s\n", time.Since(lastOperation))

	locationMap := make(map[uint64]string)

	var id uint64
	var location string
	for rows.Next() {
		err := rows.Scan(&id, &location)
		if err != nil {
			return nil, err
		}
		locationMap[id] = strings.Trim(location, "/")
	}

	util.LogInfo.Printf("create location map in getLocationMap took: %s\n", time.Since(lastOperation))
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
	lastOperation := time.Now()
	sqlString := "select " + attrCategoryID + "," + attrCategoryName + "," + attrParentID + " from " + VIPCategoryTable

	rows, err := SqlQuery(util.GetMainDB(), sqlString)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	util.LogInfo.Printf("get categories in GetCategories took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

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
	util.LogInfo.Printf("create categories in GetCategories took: %s\n", time.Since(lastOperation))

	return categories, nil
}

//get the file name stored in the disk. Keep the order of the imageID to fileName
func getRealFileNameByImageID(imageIDs []interface{}) ([]string, error) {
	lastOperation := time.Now()
	sqlString := "select " + attrID + "," + attrRawFileName + " from " + imageTable + " where " + attrID + " in (?" + strings.Repeat(",?", len(imageIDs)-1) + ")"

	rows, err := SqlQuery(db, sqlString, imageIDs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	util.LogInfo.Printf("get imaegs in getRealFileNameByImageID took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	fileMap := make(map[uint64]string)
	var fileName string
	var id uint64
	for rows.Next() {
		err := rows.Scan(&id, &fileName)
		if err != nil {
			return nil, err
		}
		fileMap[id] = fileName
	}
	util.LogInfo.Printf("create filename struct in getRealFileNameByImageID took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	if len(fileMap) != len(imageIDs) {
		return nil, errors.New("file name num doesn't match file map")
	}

	fileNameList := make([]string, 0)

	for i := 0; i < len(imageIDs); i++ {
		switch id := imageIDs[i].(type) {
		case uint64:

			if name, ok := fileMap[id]; ok {
				fileNameList = append(fileNameList, name)
			} else {
				return nil, fmt.Errorf("image id %v has no file name", id)
			}
		default:
			return nil, fmt.Errorf("image id %v(%T) type is not uint64", id, id)
		}

	}
	util.LogInfo.Printf("create imaeg map in getRealFileNameByImageID took: %s\n", time.Since(lastOperation))
	return fileNameList, nil
}

func getFileNameByImageID(imageIDs []interface{}) (map[uint64]string, error) {
	lastOperation := time.Now()
	sqlString := "select " + attrID + "," + attrFileName + " from " + imageTable + " where " + attrID + " in (?" + strings.Repeat(",?", len(imageIDs)-1) + ")"

	rows, err := SqlQuery(db, sqlString, imageIDs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	util.LogInfo.Printf("select images in getFileNameByImageID took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	fileNameMap := make(map[uint64]string, 0)
	var fileName string
	var id uint64
	for rows.Next() {
		err := rows.Scan(&id, &fileName)
		if err != nil {
			return nil, err
		}
		fileNameMap[id] = fileName
	}
	util.LogInfo.Printf("create filename map in getFileNameByImageID took: %s\n", time.Since(lastOperation))
	return fileNameMap, nil
}

func GetImageByAnswerID(answerIDs []interface{}) ([]*ImageRelation, error) {
	imageRelations := make([]*ImageRelation, 0)
	num := len(answerIDs)
	if num > 0 {
		lastOperation := time.Now()

		locationMap, err := getLocationMap()
		if err != nil {
			return nil, err
		}

		answerIDMap := make(map[uint64]*ImageRelation)
		for i := 0; i < num; i++ {
			relation := &ImageRelation{}
			relation.Info = make([]*SimpleImageInfo, 0)
			switch answerID := answerIDs[i].(type) {
			case uint64:
				relation.AnswerID = answerID
				answerIDMap[answerID] = relation
				imageRelations = append(imageRelations, relation)
			default:
				util.LogWarn.Printf("answerID has type %T instead of uint64\n", answerID)
			}
		}

		query := fmt.Sprintf("select a.%s, a.%s, a.%s, a.%s, b.%s from %s as a left join %s as b on a.%s=b.%s where b.%s in (?%s) order by b.%s",
			attrID, attrFileName, attrLocationID, attrRawFileName, attrAnswerID, imageTable, relationTable, attrID, attrImageID, attrAnswerID, strings.Repeat(",?", num-1), attrAnswerID)

		rows, err := SqlQuery(db, query, answerIDs...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		util.LogInfo.Printf("get images by answer id in GetImageByAnswerID took: %s\n", time.Since(lastOperation))
		lastOperation = time.Now()

		var imageID, locationID, answerID uint64
		var fileName string

		for rows.Next() {
			var rawFileName sql.NullString
			err = rows.Scan(&imageID, &fileName, &locationID, &rawFileName, &answerID)
			if err != nil {
				return nil, err
			}

			if relation, ok := answerIDMap[answerID]; ok {
				if locationURL, ok := locationMap[locationID]; ok {
					if rawFileName.Valid {
						u, err := url.Parse(strings.Trim(locationURL, "/") + "/" + rawFileName.String)
						if err != nil {
							util.LogError.Println(err)
							continue
						}
						imageInfo := &SimpleImageInfo{ImageID: imageID, URL: u.String()}
						relation.Info = append(relation.Info, imageInfo)
					} else {
						util.LogWarn.Printf("image id %v has no raw file name\n", imageID)
					}

				} else {
					util.LogWarn.Printf("location ID %v is not found\n", locationID)
				}
			}
		}
		util.LogInfo.Printf("create answer id map in GetImageByAnswerID took: %s\n", time.Since(lastOperation))

	}
	return imageRelations, nil

}

func createTmpStoreFolder(path string) (string, error) {
	folderName := "backupTmpPicFolder"
	fullName := path + "/" + folderName
	fi, err := os.Stat(fullName)

	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(path+"/"+folderName, 0755)
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}

	} else {
		if !fi.IsDir() {
			return "", errors.New("folder name " + folderName + " is occupied by file. Cannot create backup folder")
		}
	}

	return folderName, nil

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

func getRelationByID(ids []interface{}) (map[uint64][]uint64, error) {
	lastOperation := time.Now()
	
	sqlString := fmt.Sprintf("select %s,%s from %s where %s in(?%s)", attrImageID, attrAnswerID, relationTable, attrImageID, strings.Repeat(",?", len(ids)-1))
	rows, err := db.Query(sqlString, ids...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	util.LogInfo.Printf("select relation in getRelationByID took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	var imageID, answerID uint64
	relationMap := make(map[uint64][]uint64)
	for rows.Next() {
		err := rows.Scan(&imageID, &answerID)
		if err != nil {
			return nil, err
		}
		relationMap[imageID] = append(relationMap[imageID], answerID)
	}
	util.LogInfo.Printf("create relation map in getRelationByID took: %s\n", time.Since(lastOperation))
	return relationMap, nil
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
		fullName := prefix + name

		fi, err := os.Stat(fullName)

		if err != nil {
			if e, ok := err.(*os.PathError); ok && os.IsNotExist(e) {
				util.LogWarn.Printf("remove file %s failed. File not exist.\n", name)
				continue
			} else {
				return count, err
			}
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			return count, errors.New("Error! Try to delete folder " + fullName + " in only file allowed")
		}
		if len(name) > 0 {
			err = os.Remove(fullName)
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
		} else {
			util.LogError.Printf("Warning!! delete the empty file name! ignore")
		}

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

func CreateMediaRef(answerID int, images []int) (err error) {
	tx, err := GetTx(db)
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)
	lastOperation := time.Now()

	// read lock for images to prevent other transaction to do modification on these images
	readLockSql := fmt.Sprintf("SELECT count(id) FROM images WHERE id in (?%s) LOCK IN SHARE MODE;", strings.Repeat(",?", len(images)-1))
	var imagesParam []interface{} = make([]interface{}, len(images))
	for index, value := range images {
		imagesParam[index] = value
	}
	rows, err := tx.Query(readLockSql, imagesParam...)
	if err != nil {
		return
	}
	defer rows.Close()
	util.LogInfo.Printf("select images count in CreateMediaRef took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	var count int
	for rows.Next() {
		rows.Scan(&count)
	}

	if count != len(images) {
		return fmt.Errorf("Some media might be removed!!")
	}

	insertSql := "INSERT INTO image_answer (answer_id, image_id) VALUES"
	for index, value := range images {
		if index == 0 {
			insertSql = fmt.Sprintf("%s (%d, %d)", insertSql, answerID, value)
		} else {
			insertSql = fmt.Sprintf("%s,(%d, %d)", insertSql, answerID, value)
		}
	}
	insertSql += ";"
	_, err = tx.Exec(insertSql)
	if err != nil {
		return
	}
	util.LogInfo.Printf("insert images in CreateMediaRef took: %s\n", time.Since(lastOperation))
	return tx.Commit()
}

func DeleteMediaRef(answerID int) (err error) {
	lastOperation := time.Now()
	tx, err := GetTx(db)
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	sqlStr := fmt.Sprintf("DELETE FROM image_answer WHERE answer_id = %d", answerID)
	_, err = tx.Exec(sqlStr)
	if err != nil {
		return
	}
	util.LogInfo.Printf("delete images in DeleteMediaRef took: %s\n", time.Since(lastOperation))
	return tx.Commit()
}

func GetMetaByImageID(imageIDs []interface{}) (map[uint64]*imageMeta, error) {
	lastOperation := time.Now()
	metas := make(map[uint64]*imageMeta)
	num := len(imageIDs)
	if num == 0 {
		return metas, nil
	}
	sqlQuery := fmt.Sprintf("select %s, %s, %s from %s where %s in (?%s)", attrID, attrFileName, attrRawFileName, imageTable, attrID, strings.Repeat(",?", num-1))
	rows, err := db.Query(sqlQuery, imageIDs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	util.LogInfo.Printf("get image metadata in GetMetaByImageID took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	var fileName string
	var rawFileName string
	var id uint64
	for rows.Next() {
		err = rows.Scan(&id, &fileName, &rawFileName)
		if err != nil {
			break
		}
		metas[id] = &imageMeta{FileName: fileName, RawFileName: rawFileName}
	}
	util.LogInfo.Printf("create image metadata struct in GetMetaByImageID took: %s\n", time.Since(lastOperation))

	return metas, err

}
