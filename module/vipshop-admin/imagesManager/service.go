package imagesManager

import (
	"bytes"
	"database/sql"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func storeImage(tx *sql.Tx, fileName string, content []byte) error {
	var err error
	_, fileName, err = newImageRecord(tx, fileName, len(content))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(Volume+"/"+fileName, content, 0666)
	if err != nil {
		return err
	}

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
func newImageRecord(tx *sql.Tx, name string, size int) (uint64, string, error) {

	//db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return 0, "", errors.New("No module(" + ModuleInfo.ModuleName + ") db connection")
	}

	sql := "insert into " + imageTable + " (" + attrFileName + "," + attrLocationID + "," + attrSize + ") values (?,?,?)"

	id, _, fileName, err := inputUniqueFileName(tx, sql, name, []interface{}{name, LocalID, size})

	return uint64(id), fileName, err
}

func deleteImages(imageIDs []interface{}) (int64, error) {

	var err error

	if len(imageIDs) == 0 {
		return 0, nil
	}

	fileList, err := getFileNameByImageID(imageIDs)
	if err != nil {
		return 0, err
	}

	if len(fileList) != len(imageIDs) {
		return 0, errImageNotExist
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
		deleteFiles("", []string{Volume + "/" + folerName})

	}()

	var delFileCount int64

	delFileCount, err = deleteFiles(Volume, fileList)
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

func packageImages(imageIDs []interface{}) (*bytes.Buffer, error) {
	nameList, err := getFileNameByImageID(imageIDs)
	if err != nil {
		return nil, err
	}
	if len(nameList) != len(imageIDs) {
		return nil, errImageNotAllGet
	}

	for i := 0; i < len(nameList); i++ {
		nameList[i] = Volume + "/" + nameList[i]
	}
	var b bytes.Buffer
	err = ZipFiles(nameList, &b)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
