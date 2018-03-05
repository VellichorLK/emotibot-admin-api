package imagesManager

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/go-sql-driver/mysql"
)

func storeImage(fileName string, content []byte) error {

	id, fileName, err := newImageRecord(fileName, len(content))
	if err != nil {
		return err
	}

	//delete the record from databse in case something happened
	defer func() {
		if id != 0 {
			sql := "delete from " + imageTable + " where " + attrID + "=?"
			SqlExec(util.GetDB(ModuleInfo.ModuleName), sql, id)
		}
	}()

	err = ioutil.WriteFile(Volume+"/"+fileName, content, 0644)
	if err != nil {
		return err
	}

	id = 0
	return nil
}

func getImageList(args *getImagesArg) (*imageList, error) {

	list, err := getImageRef(args)
	if err != nil {
		return nil, err
	}

	tagMap, err := getTagMap()
	if err != nil {
		return nil, err
	}

	categoriesMap, err := GetCategories()
	if err != nil {
		return nil, err
	}

	questionInfoMap, err := getAnswerRef(list.answerIDs, tagMap, categoriesMap)
	if err != nil {
		return nil, err
	}

	for _, image := range list.Images {
		image.Refs = make([]*questionInfo, 0)
		for _, id := range image.answerID {
			qInfo, ok := questionInfoMap[id]
			if !ok {
				return nil, errors.New("No answerID " + strconv.Itoa(id))
			}
			image.Refs = append(image.Refs, qInfo)
		}
	}
	return list, nil
}

//newImageRecord would return the real file name which is inserted into db and its id
func newImageRecord(name string, size int) (uint64, string, error) {

	//db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return 0, "", errors.New("No module(" + ModuleInfo.ModuleName + ") db connection")
	}

	var id int64
	var err error
	var res sql.Result
	ext := path.Ext(name)
	baseName := name[:len(name)-len(ext)]
	counter := 0
	fileName := name
	sql := "insert into " + imageTable + " (" + attrFileName + "," + attrLocationID + "," + attrSize + ") values (?,?,?)"

	for {
		res, err = SqlExec(db, sql, fileName, LocalID, size)
		if err != nil {
			if driverErr, ok := err.(*mysql.MySQLError); ok {
				if driverErr.Number != ErDupEntry {
					return 0, "", err
				}
			} else {
				return 0, "", err
			}
		} else {
			id, err = res.LastInsertId()
			break
		}
		counter++
		fileName = baseName + "(" + strconv.Itoa(counter) + ")" + ext
	}

	return uint64(id), fileName, err
}

func deleteImages(imageIDs []interface{}) (int64, error) {

	var err error

	fileList, err := getFileNameByImageID(imageIDs)
	if err != nil {
		return 0, err
	}

	if len(fileList) != len(imageIDs) {
		return 0, errors.New("Some assigned id doesn't exist")
	}

	tx, err := GetTx(db)
	if err != nil {
		return 0, nil
	}

	defer tx.Rollback()

	sqlString := "delete from " + imageTable + " where " + attrID + " in (?" + strings.Repeat(",?", len(imageIDs)-1) + ")"
	stmt, err := tx.Prepare(sqlString)
	if err != nil {
		return 0, err
	}

	res, err := ExecStmt(stmt, imageIDs...)
	if err != nil {
		return 0, err
	}

	delRowCount, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	length := 10

	folerName, err := createBackupFolder(length, Volume)
	if err != nil {
		return 0, err
	}

	_, err = copyFiles(Volume, Volume+"/"+folerName, fileList)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err != nil {
			copyFiles(Volume+"/"+folerName, Volume, fileList)
		}
		deleteFiles([]string{Volume + "/" + folerName})

	}()

	var delFileCount int64

	delFileCount, err = deleteFiles(fileList)
	if err != nil {
		return 0, err
	}
	if delRowCount != delFileCount {
		util.LogWarn.Printf("delete images count from db(%v) is not the same from file(%v)\n", delRowCount, delFileCount)
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}
