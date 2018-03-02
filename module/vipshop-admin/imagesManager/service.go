package imagesManager

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"path"
	"strconv"

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

	db := util.GetDB(ModuleInfo.ModuleName)
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
		res, err = SqlExec(util.GetDB(ModuleInfo.ModuleName), sql, fileName, LocalID, size)
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
