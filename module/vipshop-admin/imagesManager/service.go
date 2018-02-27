package imagesManager

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"path"
	"strconv"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/go-sql-driver/mysql"
)

func storeImage(file *uploadArg) error {
	id, fileName, err := getUniqueName(file.FileName)
	if err != nil {
		return err
	}

	defer func() {
		if id != 0 {
			sql := "delete from " + imageTable + " where " + attrID + "=?"
			SqlExec(sql, id)
		}
	}()

	content, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(Volume+"/"+fileName, content, 0644)
	if err != nil {
		return err
	}

	id = 0
	return nil
}

func getUniqueName(name string) (uint64, string, error) {

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
	sql := "insert into " + imageTable + " (" + attrFileName + "," + attrLocationID + ") values (?,?)"

	for {
		res, err = SqlExec(sql, fileName, LocalID)
		if err != nil {
			if driverErr, ok := err.(*mysql.MySQLError); ok { // Now the error number is accessible directly
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
