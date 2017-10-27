package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"time"

	uuid "github.com/satori/go.uuid"
)

var args = [...]string{QFILETYPE}

var FilePrefix string

//Upload handler to store the file and start the task
func Upload(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(HXAPPID)

	if r.Method == "POST" {
		r.ParseMultipartForm(512 << 10)

		fi, err := parseParms(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fi.Appid = appid

		status, err := uploadFile(r, fi)
		if err != nil {
			http.Error(w, err.Error(), status)
			return
		}

		status, err = sendTask(fi, appid)
		if err != nil {
			http.Error(w, err.Error(), status)
			return
		}

		res, status, err := packageUploadReturn(fi)
		if err != nil {
			http.Error(w, err.Error(), status)
			return
		}

		contentType := "application/json; charset=utf-8"

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(res)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func parseParms(r *http.Request) (*FileInfo, error) {

	f := new(FileInfo)

	//check the required args
	for i := 0; i < len(args); i++ {
		if r.FormValue(args[i]) == "" {
			stringErr := "Need " + args[i] + " argument"
			return nil, errors.New(stringErr)
		}
	}

	f.Checksum = r.FormValue(QCHECKSUM)
	if f.Checksum != "" {
		if !regexp.MustCompile("^[a-fA-F0-9]+$").MatchString(f.Checksum) || len(f.Checksum) != 32 {
			return nil, errors.New(NCHECKSUM + "can only contain a-f,0-9 with length 32")
		}
	}
	f.FileType = r.FormValue(QFILETYPE)

	if r.FormValue(QDURATION) != "" {
		durationInt, err := strconv.Atoi(r.FormValue(QDURATION))
		if err != nil {
			return nil, errors.New("Wrong type of " + QDURATION)
		} else if durationInt < 0 {
			return nil, errors.New(QDURATION + " < 0")
		}
		f.Duration = uint32(durationInt)
	}

	now := time.Now()
	if r.FormValue(QFILET) == "" {
		f.CreateTime = uint64(time.Now().Unix())
	} else {
		createTimeInt64, err := strconv.ParseInt(r.FormValue(QFILET), 10, 64)
		if err != nil {
			return nil, errors.New("Wrong type of " + QFILET)
		} else if createTimeInt64 < 0 {
			return nil, errors.New(QFILET + " < 0")
		} else if createTimeInt64 > now.Unix() {
			return nil, errors.New("do not do future time (" + QFILET + "). time traveler. ")
		}
		f.CreateTime = uint64(createTimeInt64)
	}

	f.Priority = DEFAULTPRIORITY
	/*
		if r.FormValue(QPRIORITY) != "" {
			priorityInt8, err := strconv.ParseInt(r.FormValue(QPRIORITY), 10, 8)
			if err != nil {
				return nil, errors.New("Wrong type of " + QPRIORITY)
			} else if priorityInt8 < 0 || priorityInt8 > 4 {
				return nil, errors.New(QPRIORITY + " out of range 0~4")
			}
			f.Priority = uint8(priorityInt8)
		} else {
			f.Priority = DEFAULTPRIORITY
		}
	*/

	if r.FormValue(QTAGS) != "" {
		tags := strings.Split(r.FormValue(QTAGS), ",")
		bound := len(tags)
		if bound > LIMITUSERTAGS {
			bound = LIMITUSERTAGS
		}
		f.Tags = tags[:bound]

		if hasDupTags(f.Tags) {
			return nil, errors.New("has duplicate tags")
		}

	} else {
		f.Tags = make([]string, 0)
	}

	if r.FormValue(NUSRCOL) != "" {
		ucs := strings.Split(r.FormValue(NUSRCOL), ",")
		for _, v := range ucs {
			uc := strings.Split(v, "==")
			if len(uc) != 2 {
				return nil, errors.New("wrong format of user_column")
			}
			owner, ok := DefaulUsrField.FieldOwner[uc[0]]
			if !ok || strings.Compare(owner, r.Header.Get(HXAPPID)) != 0 {
				return nil, errors.New("no colum id " + uc[0])
			}

			nameInterface, ok := DefaulUsrField.FieldNameMap.Load(uc[0])
			if !ok {
				return nil, errors.New("no colum id " + uc[0] + " name")
			}

			if !checkSelectableVal(uc[0], uc[1]) {
				return nil, errors.New("value " + uc[1] + " can't be set.")
			}

			name := nameInterface.(string)
			cv := &ColumnValue{ColID: uc[0], Value: uc[1], Field: name}
			f.UsrColumn = append(f.UsrColumn, cv)
		}
	} else {
		dvsInterface, ok := DefaulUsrField.DefaultValue.Load(r.Header.Get(HXAPPID))
		if ok {
			dvs := dvsInterface.([]*DefaultValue)
			for _, dv := range dvs {

				cv := &ColumnValue{ColID: dv.ColID, Value: dv.ColValue, Field: dv.ColName}
				f.UsrColumn = append(f.UsrColumn, cv)
			}

		}

	}

	return f, nil
}

func hasDupTags(tags []string) bool {
	values := make(map[string]bool)
	if tags != nil {
		for _, v := range tags {
			if _, ok := values[v]; ok {
				return true
			}
			values[v] = true
		}
	}
	return false
}

func uploadFile(r *http.Request, fi *FileInfo) (int, error) {
	file, handler, err := r.FormFile(QFILE)
	if err != nil {
		return http.StatusBadRequest, err
	}
	defer file.Close()

	//check md5sum
	if fi.Checksum != "" {
		md5h := md5.New()
		io.Copy(md5h, file)
		checksum := hex.EncodeToString(md5h.Sum(nil))
		if strings.EqualFold(fi.Checksum, checksum) {
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

	now := time.Now()
	fi.UPTime = now.Unix()
	fi.UploadTime = fi.UPTime
	err := updateDatabase(fi)
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
	task.QueueN = QUEUEMAP["taskQueue"].Name
	//task.QueueN = "ecovacasQueue"

	err = goTask(task)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func pacakgeTask(fi *FileInfo) ([]byte, error) {
	task := &TaskBlock{Path: fi.Path, File: fi.UFileName}
	encodeTask, err := json.Marshal(task)
	if err != nil {
		log.Printf("path:%s, file:%s, extension:%s\n", fi.Path, fi.FileName, fi.FileType)
		return nil, err
	}
	return encodeTask, nil
	//return string(encodeTask), nil
}

func updateDatabase(fi *FileInfo) error {
	err := InsertFileRecord(fi)
	if err != nil {
		return err
	}
	return InsertUserDefinedTags(fi.ID, fi.Tags)
}

func goTaskOnFail(sid string) {

	//remove database record
	id, _ := strconv.ParseUint(sid, 10, 64)
	_, err := DeleteFileRecord(id)
	if err != nil {
		log.Println(err)
	}
	_, err = ExecuteSQL(DeleteFileRowSQL, id)
	if err != nil {
		log.Println(err)
	}
	_, err = DeleteTag(id)
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
