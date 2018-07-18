package imagesManager

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/satori/go.uuid"
)

func storeImage(tx *sql.Tx, fileName string, content []byte) (string, error) {
	var err error

	u1, err := uuid.NewV4()
	if err != nil {
		return fileName, err
	}

	//image name store in the disk
	rawFileName := hex.EncodeToString(u1[:]) + path.Ext(fileName)

	_, fileName, err = newImageRecord(tx, fileName, len(content), rawFileName)
	if err != nil {
		return fileName, err
	}

	err = ioutil.WriteFile(Volume+"/"+rawFileName, content, 0644)
	if err != nil {
		return fileName, err
	}

	return fileName, nil
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
			if ok {
				image.Refs = append(image.Refs, qInfo)
			}

		}
	}
	return list, nil
}

//newImageRecord would return the real file name which is inserted into db and its id
func newImageRecord(tx *sql.Tx, name string, size int, rawFileName string) (uint64, string, error) {

	if db == nil {
		return 0, "", errors.New("No module(" + ModuleInfo.ModuleName + ") db connection")
	}

	insertSQL := fmt.Sprintf("insert into %s (%s,%s,%s,%s) values (?,?,?,?)", imageTable, attrFileName, attrLocationID, attrSize, attrRawFileName)

	id, _, fileName, err := inputUniqueFileName(tx, insertSQL, name, []interface{}{name, LocalID, size, rawFileName})

	return uint64(id), fileName, err
}

func deleteImages(imageIDs []interface{}) (int64, error) {

	var err error

	if len(imageIDs) == 0 {
		return 0, nil
	}

	fileList, err := getRealFileNameByImageID(imageIDs)
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

	defer stmt.Close()

	res, err := ExecStmt(stmt, imageIDs...)
	if err != nil {
		return 0, err
	}

	delRowCount, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	folderName, err := createTmpStoreFolder(Volume)
	if err != nil {
		return 0, err
	}

	_, err = copyFiles(Volume, Volume+"/"+folderName, fileList)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err != nil {
			copyFiles(Volume+"/"+folderName, Volume, fileList)
		}
		deleteFiles(Volume+"/"+folderName, fileList)
	}()

	//delete files
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

	return delRowCount, nil
}

func packageImages(imageIDs []interface{}) (*bytes.Buffer, error) {
	nameMap, err := getFileNameByImageID(imageIDs)
	if err != nil {
		return nil, err
	}

	rawFileName, err := getRealFileNameByImageID(imageIDs)
	if err != nil {
		return nil, err
	}

	if len(nameMap) != len(imageIDs) || len(rawFileName) != len(imageIDs) {
		return nil, errImageNotAllGet
	}
	realFileLocation := make([]string, len(imageIDs))
	nameList := make([]string, len(imageIDs))

	for i := 0; i < len(imageIDs); i++ {
		switch id := imageIDs[i].(type) {
		case uint64:
			if file, ok := nameMap[id]; ok {
				realFileLocation[i] = Volume + "/" + rawFileName[i]
				nameList[i] = file
			} else {
				return nil, fmt.Errorf("image id %v has no file name", id)
			}

		default:
			return nil, fmt.Errorf("image id %v(%T) type doesn't match", id, id)
		}
	}
	var b bytes.Buffer
	err = ZipFiles(nameList, realFileLocation, &b)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func getAnswerImage(answerIDs []uint64) ([]*ImageRelation, error) {
	ids := make([]interface{}, len(answerIDs))
	for i := 0; i < len(answerIDs); i++ {
		ids[i] = answerIDs[i]
	}
	return GetImageByAnswerID(ids)
}
