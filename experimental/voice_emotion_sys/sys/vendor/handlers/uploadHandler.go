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

	"time"

	uuid "github.com/satori/go.uuid"
)

var args = [...]string{NFILETYPE, NDURATION, NFILET, NCHECKSUM, NTAG, NAPPID}
var optional = [...]string{NCHECKSUM, NTAG, NAPPID}
var FilePrefix string

//Upload handler to store the file and start the task
func Upload(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(NAPPID)
	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	durationInt, err := strconv.Atoi(r.FormValue(NDURATION))
	if err != nil {
		return nil, errors.New("Wrong type of " + NDURATION)
	} else if durationInt < 0 {
		return nil, errors.New(NDURATION + " < 0")
	}
	f.Duration = uint32(durationInt)

	now := time.Now()
	createTimeInt64, err := strconv.ParseInt(r.FormValue(NFILET), 10, 64)
	if err != nil {
		return nil, errors.New("Wrong type of " + NFILET)
	} else if createTimeInt64 < 0 {
		return nil, errors.New(NFILET + " < 0")
	} else if createTimeInt64 > now.Unix() {
		return nil, errors.New("do not do future time (" + NFILET + "). time traveler. ")

	}
	f.CreateTime = uint64(createTimeInt64)
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

	//create string uuid
	uuid := uuid.NewV4()
	corrID := hex.EncodeToString(uuid[:])
	fi.FileID = corrID

	fi.UFileName = fi.FileID + "-" + handler.Filename
	fi.FileName = handler.Filename
	f, err := os.OpenFile("./upload_file/"+fi.Appid+"/"+fi.UFileName, os.O_WRONLY|os.O_CREATE, 0644)
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

	now := time.Now()
	fi.UPTime = now.Unix()
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
	//task.QueueN = QUEUEMAP[appid].Name
	task.QueueN = "ecovacasQueue"

	err = goTask(task)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func pacakgeTask(fi *FileInfo) (string, error) {
	task := &TaskBlock{Path: fi.Path, File: fi.UFileName}
	encodeTask, err := json.Marshal(task)
	if err != nil {
		log.Printf("path:%s, file:%s, extension:%s\n", fi.Path, fi.FileName, fi.FileType)
		return "", err
	}
	return string(encodeTask), nil
}

func updateDatabase(fi *FileInfo) error {
	return InsertFileRecord(fi)
}

func goTaskOnFail(sid string) {
	id, _ := strconv.ParseUint(sid, 10, 64)
	err := DeleteFileRecord(id)
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
