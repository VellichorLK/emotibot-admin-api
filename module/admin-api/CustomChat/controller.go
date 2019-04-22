package CustomChat

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/audit"
	"emotibot.com/emotigo/module/admin-api/util/localemsg"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/admin-api/util/validate"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func initDao() {
	dao = customChatDao{}
}

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "customchat",
		EntryPoints: append(
			[]util.EntryPoint{
				util.NewEntryPointWithVer("POST", "import/question", []string{"view"}, handleImportCustomChatQuestion, 1),
				util.NewEntryPointWithVer("POST", "import/extend", []string{"view"}, handleImportCustomChatExtend, 1),
				util.NewEntryPointWithVer("GET", "export/question", []string{"export"}, handleExportCustomChatQuestion, 1),
			}),
	}

	initDao()
}

func handleImportCustomChatQuestion(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	var err AdminErrors.AdminError
	var auditMsg bytes.Buffer

	defer func() {
		retVal := 0
		if err == nil {
			retVal = 1
		} else {
			auditMsg.WriteString(":")
			auditMsg.WriteString(err.Error())
		}

		audit.AddAuditFromRequestAuto(r, auditMsg.String(), retVal)
		util.Return(w, err, auditMsg.String())
	}()
	auditMsg.WriteString(util.Msg["UploadCustomChatQuestion"])

	file, info, ioErr := r.FormFile("file")
	if ioErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, "")
		return
	}

	var buffer bytes.Buffer
	_, ioErr = io.Copy(&buffer, file)
	if ioErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, "")
		return
	}
	auditMsg.WriteString(info.Filename)

	customQuestions, parseErr := ParseImportQuestionFile(buffer.Bytes(), locale)

	if parseErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, parseErr.Error())
		return
	}
	auditMsg.WriteString(fmt.Sprintf(util.Msg["UploadCustomChatInfoTpl"], len(customQuestions)))

	err = UpdateLatestCustomChatQuestions(appid, customQuestions)
	if err != nil {
		return
	}

	go SyncCustomChat(appid, false)

	return
}

func handleImportCustomChatExtend(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	var err AdminErrors.AdminError
	var auditMsg bytes.Buffer

	defer func() {
		retVal := 0
		if err == nil {
			retVal = 1
		} else {
			auditMsg.WriteString(":")
			auditMsg.WriteString(err.Error())
		}

		audit.AddAuditFromRequestAuto(r, auditMsg.String(), retVal)
		util.Return(w, err, auditMsg.String())
	}()
	auditMsg.WriteString(util.Msg["UploadCustomChatExtend"])

	file, info, ioErr := r.FormFile("file")
	if ioErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, "")
		return
	}

	var buffer bytes.Buffer
	_, ioErr = io.Copy(&buffer, file)
	if ioErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, "")
		return
	}
	auditMsg.WriteString(info.Filename)

	questions, parseErr := ParseImportExtendFile(buffer.Bytes(), locale)

	if parseErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, parseErr.Error())
		return
	}
	auditMsg.WriteString(fmt.Sprintf(util.Msg["UploadCustomChatExtendInfoTpl"], len(questions)))

	err = UpdateLatestCustomChatExtends(appid, questions)
	if err != nil {
		return
	}

	go SyncCustomChat(appid, false)

	return
}

func handleExportCustomChatQuestion(w http.ResponseWriter, r *http.Request) {

	//appid := r.URL.Query().Get("appid")
	appid := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	var err AdminErrors.AdminError
	var ret []byte
	var auditMsg bytes.Buffer

	defer func() {
		retVal := 0
		if err == nil {
			retVal = 1
			now := time.Now()
			filename := fmt.Sprintf("custom_chat_%d%02d%02d_%02d%02d%02d.xlsx",
				now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
			util.ReturnFile(w, filename, ret)
			auditMsg.WriteString(":")
			auditMsg.WriteString(filename)
		} else {
			auditMsg.WriteString(":")
			auditMsg.WriteString(err.Error())
			util.Return(w, err, nil)
		}
		audit.AddAuditFromRequestAuto(r, auditMsg.String(), retVal)
	}()
	auditMsg.WriteString(localemsg.Get(locale, "CustomChatExport"))

	if !validate.IsValidAppID(appid) {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, "APPID")
		return
	}

	ret, err = GetExportCustomChat(appid, locale)
	return
}

//func Test(w http.ResponseWriter, r *http.Request)  {
//	appid := requestheader.GetAppID(r)
//	SyncCustomChat(appid, true)
//}
