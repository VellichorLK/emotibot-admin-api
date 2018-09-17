package handlers

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"time"

	uuid "github.com/satori/go.uuid"
)

//var args = [...]string{NFILETYPE, NDURATION, NFILET, NCHECKSUM, NTAG, NAPPID}
var args = [...]string{NFILETYPE, NCHECKSUM, NTAG, NAPPID}
var optional = [...]string{NCHECKSUM, NTAG, NAPPID}
var FilePrefix string

//Upload handler to store the file and start the task
func Upload(w http.ResponseWriter, r *http.Request) {

	var logStatus int

	var fi *FileInfo
	defer func() {

		var appid, fileName, tag1, tag2 string
		var createTime uint64
		if fi != nil {
			appid = fi.Appid
			fileName = fi.FileName
			tag1 = fi.Tag
			tag2 = fi.Tag2
			createTime = fi.CreateTime
		}

		fmt.Printf("[%s] ret code:%d, appid:%s, file_name:%s, tag1:%s, tag2:%s, create_time:%s   ",
			time.Now().Format(time.RFC3339), logStatus, appid, fileName, tag1, tag2, time.Unix(int64(createTime), 0).Format(time.RFC3339))
	}()

	appid := r.Header.Get(NUAPPID)
	if appid == "" {
		logStatus = http.StatusUnauthorized
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" {
		r.ParseMultipartForm(512 << 10)
		var err error

		fi, err = parseParms(r)
		if err != nil {
			logStatus = http.StatusBadRequest
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fi.Appid = appid

		status, err := uploadFile(r, fi)
		if err != nil {
			logStatus = status
			http.Error(w, err.Error(), status)
			return
		}

		status, err = sendTask(fi, appid)
		if err != nil {
			logStatus = status
			http.Error(w, err.Error(), status)
			return
		}

		res, status, err := packageUploadReturn(fi)
		if err != nil {
			logStatus = status
			http.Error(w, err.Error(), status)
			return
		}

		contentType := "application/json; charset=utf-8"

		logStatus = http.StatusOK

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(res)

	} else {
		logStatus = http.StatusMethodNotAllowed
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func parseParms(r *http.Request) (*FileInfo, error) {

	numArgs := len(args)
	numOption := len(optional)
	f := new(FileInfo)

	//check the required args
	for i := 0; i < (numArgs - numOption); i++ {
		if r.FormValue(args[i]) == "" {
			stringErr := "Need " + args[i] + " argument"
			return nil, errors.New(stringErr)
		}
	}

	f.FileType = r.FormValue(NFILETYPE)
	f.Checksum = r.FormValue(NCHECKSUM)
	f.Tag = r.FormValue(NTAG)
	f.Tag2 = r.FormValue(NTAG2)
	if f.Tag != "" {
		if len(f.Tag) > LIMITTAGLEN {
			return nil, errors.New(NTAG + " over limit " + strconv.Itoa(LIMITTAGLEN))
		} else if !regexp.MustCompile("^[-_.a-zA-Z0-9]+$").MatchString(f.Tag) {
			return nil, errors.New(NTAG + " can only cantain a-z,A-Z,0-9,-,_,.character")
		}
	}

	// make duration as optional field
	f.Duration = uint32(0)
	durationInt, err := strconv.Atoi(r.FormValue(NDURATION))
	if err != nil {
		// return nil, errors.New("Wrong type of " + NDURATION)
	} else if durationInt < 0 {
		return nil, errors.New(NDURATION + " < 0")
	} else {
		// set if given
		f.Duration = uint32(durationInt)
	}

	now := time.Now()
	createTimeInt64, err := strconv.ParseInt(r.FormValue(NFILET), 10, 64)
	if err != nil || createTimeInt64 < 0 {
		createTimeInt64 = now.Unix()
	}
	f.CreateTime = uint64(createTimeInt64)
	f.UploadTime = uint64(now.Unix())
	f.UPTime = now.Unix()

	/*
		if err != nil {
			return nil, errors.New("Wrong type of " + NFILET)
		} else if createTimeInt64 < 0 {
			return nil, errors.New(NFILET + " < 0")
		} else if createTimeInt64 > now.Unix() {
			return nil, errors.New("do not do future time (" + NFILET + "). time traveler. ")

		}
	*/
	f.Priority = DEFAULTPRIORITY
	/*
		if r.FormValue(NPRIORITY) != "" {
			priorityInt8, err := strconv.ParseInt(r.FormValue(NPRIORITY), 10, 8)
			if err != nil {
				return nil, errors.New("Wrong type of " + NPRIORITY)
			} else if priorityInt8 < 0 || priorityInt8 > 4 {
				return nil, errors.New(NPRIORITY + " out of range 0~4")
			}
			f.Priority = uint8(priorityInt8)
		} else {
			f.Priority = DEFAULTPRIORITY
		}
	*/

	return f, nil
}

func uploadFile(r *http.Request, fi *FileInfo) (int, error) {
	file, handler, err := r.FormFile(NFILE)
	if err != nil {
		return http.StatusBadRequest, err
	}
	defer file.Close()

	//check md5sum
	if fi.Checksum != "" {
		md5h := md5.New()
		io.Copy(md5h, file)
		checksum := hex.EncodeToString(md5h.Sum(nil))
		if fi.Checksum != checksum {
			return http.StatusBadRequest, errors.New("checksum mismatch")
		}
		file.Seek(0, 0)
	}

	filePrefix := "./upload_file/" + fi.Appid

	//check appid folder exist
	_, err = os.Stat(filePrefix)
	if os.IsNotExist(err) {
		err = os.Mkdir(filePrefix, 0755)
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("[Warning] create appid folder %s\n", filePrefix)
		}
	}

	//create string uuid
	uuid := uuid.NewV4()
	corrID := hex.EncodeToString(uuid[:])
	fi.FileID = corrID

	fi.UFileName = fi.FileID + "-" + handler.Filename
	fi.FileName = handler.Filename
	f, err := os.OpenFile(filePrefix+"/"+fi.UFileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer f.Close()

	fileSize, err := io.Copy(f, file)
	if err != nil {
		log.Println(err)
		return http.StatusInternalServerError, err
	}
	fi.Size = uint64(fileSize)
	fi.Path = FilePrefix + "/" + fi.Appid
	return http.StatusOK, nil

}

func sendTask(fi *FileInfo, appid string) (int, error) {

	/*
		if QUEUEMAP[appid].Name == "" {
			return http.StatusBadRequest, errors.New("appid has no assigned service queue")
		}
	*/

	//now := time.Now()
	//fi.UPTime = now.Unix()

	tx, err := db.Begin()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer tx.Rollback()

	err = updateDatabase(tx, fi)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	taskStr, err := pacakgeTask(fi)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	task := new(Task)
	task.PackagedTask = taskStr
	task.FileInfo = fi
	//task.QueueN = QUEUEMAP[appid].Name
	task.QueueN = "ecovacasQueue"

	err = goTask(task)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	tx.Commit()

	return http.StatusOK, nil
}

func pacakgeTask(fi *FileInfo) (string, error) {
	task := &TaskBlock{Path: fi.Path, File: fi.UFileName, Appid: fi.Appid}
	encodeTask, err := json.Marshal(task)
	if err != nil {
		log.Printf("path:%s, file:%s, extension:%s\n", fi.Path, fi.FileName, fi.FileType)
		return "", err
	}
	return string(encodeTask), nil
}

func updateDatabase(tx *sql.Tx, fi *FileInfo) error {
	return InsertFileRecord(tx, fi)
}

func goTaskOnFail(sid string) {

	//remove database record
	err := ExecSQL(DeleteFileRowSQL, sid)
	if err != nil {
		log.Println(err)
	}
	err = ExecSQL(DeleteUsrFieldValueSQL, sid)
	if err != nil {
		log.Println(err)
	}
}

func goTask(task *Task) error {

	select {
	case TaskQueue <- task:
		break
	case <-time.After(2 * time.Second):
		goTaskOnFail(task.FileInfo.ID)
		return errors.New("Push task to channel failed. Please retry later")
	}
	select {
	case ok := <-RelyQueue:
		if ok {
			return nil
		}
		goTaskOnFail(task.FileInfo.ID)
		return errors.New("Push task to queue failed. Please retry later")
	case <-time.After(2 * time.Second):
		return errors.New("Get channel result failed after 2 second")
	}

}

func packageUploadReturn(fi *FileInfo) ([]byte, int, error) {

	fi.Channels = make([]*ChannelResult, 0, 0)

	encodeRes, err := json.Marshal(fi.ReturnBlock)
	if err != nil {
		log.Println(err)
		return nil, http.StatusInternalServerError, errors.New("Internal server error")
	}
	return encodeRes, http.StatusOK, nil
}
