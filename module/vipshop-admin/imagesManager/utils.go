package imagesManager

import (
	"archive/zip"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func reverseSlice(name []string) {
	for i := 0; i < len(name)/2; i++ {
		j := len(name) - i - 1
		name[i], name[j] = name[j], name[i]
	}
}

func getImagesParams(params map[string]string) (*getImagesArg, error) {
	listArgs := &getImagesArg{Order: attrID, Page: 0, Limit: 12}

	for k, v := range params {
		switch k {
		case ORDER:
			switch v {
			case valID:
				listArgs.Order = attrID
			case valName:
				listArgs.Order = attrFileName
			case valTime:
				listArgs.Order = attrLatestUpdate
			default:
				return nil, errors.New("Unknown parameter value " + v)
			}
		case PAGE:
			page, err := strconv.ParseInt(v, 10, 64)
			if err != nil || page < 0 {
				return nil, errors.New(k + "(" + v + ")" + " is not numeric or negative")
			}
			listArgs.Page = page
		case LIMIT:
			limit, err := strconv.ParseInt(v, 10, 64)
			if err != nil || limit <= 0 {
				return nil, errors.New(k + "(" + v + ")" + " is not numeric or <=0")
			}
			listArgs.Limit = limit
		case KEYWORD:
			listArgs.Keyword = v
		default:
			return nil, errors.New("Unknown parameter " + k)
		}
	}

	return listArgs, nil
}

func GetFullCategory(categories map[int]*Category, categoryID int) ([]string, error) {

	if categories == nil {
		return nil, errors.New("categories map is nil")
	}

	const MAXLEVEL = 5
	levels := make([]string, 0, MAXLEVEL)
	for i := 0; i < MAXLEVEL; i++ {
		c, ok := categories[categoryID]
		if !ok {
			return nil, errors.New("No categoryID " + strconv.Itoa(categoryID))
		}
		levels = append(levels, c.Name)
		if c.ParentID == 0 {
			break
		}
		categoryID = c.ParentID
	}

	levels = levels[:len(levels)]
	reverseSlice(levels)
	return levels, nil
}

var src = rand.NewSource(time.Now().UnixNano())

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GetUniqueString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

type mockedInfo struct {
	real os.FileInfo
}

// underlying data source (can return nil)
func (m mockedInfo) Sys() interface{} {
	return m.real.Sys()
}

//ZipFiles files is the real name store in the disk, fileName is the output file name you want
func ZipFiles(files []string, fileName []string, target io.Writer) error {

	zipWriter := zip.NewWriter(target)
	defer zipWriter.Close()

	for i := 0; i < len(fileName); i++ {
		file, err := os.Open(fileName[i])
		if err != nil {
			return err
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = files[i]

		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}
	}

	return nil
}

//Md5Uint64 do md5sum with uint64 input
func Md5Uint64(id uint64) string {
	h := md5.New()
	fmt.Fprint(h, id)
	encode := h.Sum(nil)
	return hex.EncodeToString(encode)
}

//get the image name store in the disk by id
func getImageName(id uint64) (string, error) {

	getImageNameSQL := fmt.Sprintf("select %s from %s where %s=?", attrRawFileName, imageTable, attrID)

	var fileName sql.NullString
	err := db.QueryRow(getImageNameSQL, id).Scan(&fileName)
	if err != nil {
		return "", err
	}

	if !fileName.Valid {
		return "", errors.New("file name is null")
	}

	return fileName.String, nil
}
