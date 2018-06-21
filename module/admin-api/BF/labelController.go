package BF

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
)

func handleGetCmds(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	var err error
	retCode := ApiError.SUCCESS
	defer func() {
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, err.Error()), ApiError.GetHttpStatus(retCode))
		}
	}()
	appid := util.GetAppID(r)

	cmds, err := GetCmds(appid)
	if err != nil {
		retCode = ApiError.DB_ERROR
	} else {
		retObj = cmds
	}
}
func handleGetCmd(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	var err error
	retCode := ApiError.SUCCESS
	defer func() {
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, err.Error()), ApiError.GetHttpStatus(retCode))
		}
	}()
	appid := util.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		retCode = ApiError.REQUEST_ERROR
		retObj = err.Error()
		return
	}

	cmd, err := GetCmd(appid, id)
	if cmd == nil {
		retCode = ApiError.NOT_FOUND_ERROR
		err = ApiError.ErrNotFound
		return
	} else if err != nil {
		retCode = ApiError.DB_ERROR
		return
	}

	retObj = cmd
	return
}
func handleUpdateCmd(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	var err error
	appid := util.GetAppID(r)
	retCode := ApiError.SUCCESS
	defer func() {
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, err.Error()), ApiError.GetHttpStatus(retCode))
		}
	}()

	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		retCode = ApiError.REQUEST_ERROR
		err = ApiError.GenBadRequestError("ID")
		return
	}

	origCmd, err := GetCmd(appid, id)
	if origCmd == nil {
		retCode = ApiError.NOT_FOUND_ERROR
		err = ApiError.ErrNotFound
		return
	}
	if err != nil {
		retCode = ApiError.IO_ERROR
		return
	}

	cmd, err := parseCmdFromRequest(r)
	if err != nil {
		retCode = ApiError.REQUEST_ERROR
		return
	}

	retCode, err = UpdateCmd(appid, id, cmd)
	if err != nil {
		return
	}
	cmd.ID = id
	retObj = cmd
	go util.ConsulUpdateCmd(appid)
}
func handleAddCmd(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	var err error
	appid := util.GetAppID(r)
	retCode := ApiError.SUCCESS
	defer func() {
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, err.Error()), ApiError.GetHttpStatus(retCode))
		}
	}()

	cmd, err := parseCmdFromRequest(r)
	if err != nil {
		retCode = ApiError.REQUEST_ERROR
		return
	}
	cid, err := strconv.Atoi(r.FormValue("cid"))
	if err != nil {
		err = ApiError.GenBadRequestError(util.Msg["CmdParentID"])
		retCode = ApiError.REQUEST_ERROR
		return
	}

	id, retCode, err := AddCmd(appid, cmd, cid)
	if err != nil {
		return
	}
	cmd.ID = id
	retObj = cmd
	go util.ConsulUpdateCmd(appid)
}
func handleDeleteCmd(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	var err error
	appid := util.GetAppID(r)
	retCode := ApiError.SUCCESS
	defer func() {
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, err.Error()), ApiError.GetHttpStatus(retCode))
		}
	}()
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		retCode = ApiError.REQUEST_ERROR
		retObj = err.Error()
		return
	}

	err = DeleteCmd(appid, id)
	if err != nil {
		retCode = ApiError.DB_ERROR
	}
	if err == nil {
		go util.ConsulUpdateCmd(appid)
	}
}
func handleGetCmdsOfLabel(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	status, retCode := http.StatusOK, ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()
	appid := util.GetAppID(r)
	labelID, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		status, retCode = http.StatusBadRequest, ApiError.REQUEST_ERROR
		return
	}

	cmds, err := GetCmdsOfLabel(appid, labelID)
	if err != nil {
		status, retCode = http.StatusInternalServerError, ApiError.DB_ERROR
		retObj = err.Error()
	} else {
		retObj = cmds
	}
}
func parseCmdFromRequest(r *http.Request) (cmd *Cmd, err error) {
	err = r.ParseForm()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			util.LogInfo.Printf("Parse cmd fail: %s\n", err.Error())
		}
	}()

	ret := Cmd{}
	ret.Name = r.FormValue("name")
	ret.Answer = r.FormValue("answer")
	ret.Status = r.FormValue("status") != "0"
	begin, err := time.Parse(time.RFC3339, r.FormValue("begin_time"))
	if err != nil {
		ret.Begin = nil
	} else {
		ret.Begin = &begin
	}
	end, err := time.Parse(time.RFC3339, r.FormValue("end_time"))
	if err != nil {
		ret.End = nil
	} else {
		ret.End = &end
	}

	target, err := strconv.Atoi(r.FormValue("target"))
	if err != nil {
		err = errors.New("Invalid target")
		return
	}
	rtype, err := strconv.Atoi(r.FormValue("response_type"))
	if err != nil {
		err = errors.New("Invalid response type")
		return
	}

	if target > ret.Target.Max() || target < 0 {
		err = errors.New("Invalid target")
		return
	}
	if rtype > ret.Type.Max() || rtype < 0 {
		err = errors.New("Invalid response type")
		return
	}
	ret.Target = CmdTarget(target)
	ret.Type = ResponseType(rtype)

	ruleStr := r.FormValue("rule")
	ruleContents := []*CmdContent{}
	err = json.Unmarshal([]byte(ruleStr), &ruleContents)
	if err != nil {
		err = fmt.Errorf("Invalid rule content: %s", err.Error())
		return
	}
	for i, r := range ruleContents {
		if !r.IsValid() {
			err = fmt.Errorf("rule content error of rule %d", i+1)
			return
		}
	}
	ret.Rule = ruleContents

	labelsStr := r.FormValue("labels")
	labelIDs := []string{}
	err = json.Unmarshal([]byte(labelsStr), &labelIDs)
	if err != nil {
		return
	}
	existedLabel := map[string]bool{}
	for _, id := range labelIDs {
		if _, ok := existedLabel[id]; !ok {
			ret.LinkLabel = append(ret.LinkLabel, id)
			existedLabel[id] = true
		}
	}

	cmd = &ret
	return
}
func handleGetLabelsOfCmd(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	var err error
	retCode := ApiError.SUCCESS
	defer func() {
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, err.Error()), ApiError.GetHttpStatus(retCode))
		}
	}()
	appid := util.GetAppID(r)
	cmdID, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		retCode = ApiError.REQUEST_ERROR
		return
	}

	labels, err := GetLabelsOfCmd(appid, cmdID)
	if err != nil {
		retCode = ApiError.DB_ERROR
	} else {
		retObj = labels
	}
}
func handleGetCmdClass(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	var err error
	retCode := ApiError.SUCCESS
	defer func() {
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, err.Error()), ApiError.GetHttpStatus(retCode))
		}
	}()
	appid := util.GetAppID(r)
	classID, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		retCode, err = ApiError.REQUEST_ERROR, ApiError.GenBadRequestError("ID")
		return
	}

	retObj, retCode, err = GetCmdClass(appid, classID)
	return
}
func handleDeleteCmdClass(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	var err error
	retCode := ApiError.SUCCESS
	defer func() {
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, err.Error()), ApiError.GetHttpStatus(retCode))
		}
	}()
	appid := util.GetAppID(r)
	classID, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		err = ApiError.GenBadRequestError(util.Msg["CmdParentID"])
		retCode = ApiError.REQUEST_ERROR
		return
	}

	err = DeleteCmdClass(appid, classID)
	if err != nil {
		retCode = ApiError.DB_ERROR
		return
	}
}
func handleAddCmdClass(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	var err error
	retCode := ApiError.SUCCESS
	defer func() {
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, err.Error()), ApiError.GetHttpStatus(retCode))
		}
	}()
	appid := util.GetAppID(r)
	className := r.FormValue("name")
	if strings.TrimSpace(className) == "" {
		retCode, err = ApiError.REQUEST_ERROR, ApiError.GenBadRequestError(util.Msg["CmdClassName"])
		return
	}

	// class layer is only one for now
	// pid, err := strconv.Atoi(r.FormValue("pid"))
	// if err != nil {
	// 	retCode, retObj = ApiError.REQUEST_ERROR, fmt.Sprintf("get pid fail: %s", err.Error())
	// 	status = http.StatusBadRequest
	// 	return
	// }
	// class, err = GetCmdClass(appid, pid)
	// if err != nil {
	// 	retCode, retObj = ApiError.DB_ERROR, fmt.Sprintf("get parent class fail")
	// }

	var pid *int
	classID, retCode, err := AddCmdClass(appid, pid, className)
	if err != nil {
		return
	}
	retObj, retCode, err = GetCmdClass(appid, classID)
}
func handleUpdateCmdClass(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	var err error
	retCode := ApiError.SUCCESS
	defer func() {
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, err.Error()), ApiError.GetHttpStatus(retCode))
		}
	}()
	appid := util.GetAppID(r)
	classID, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		err = ApiError.GenBadRequestError(util.Msg["CmdParentID"])
		retCode = ApiError.REQUEST_ERROR
		return
	}

	newClassName := r.FormValue("name")
	if strings.TrimSpace(newClassName) == "" {
		retCode, err = ApiError.REQUEST_ERROR, ApiError.GenBadRequestError(util.Msg["CmdClassName"])
		return
	}

	retCode, err = UpdateCmdClass(appid, classID, newClassName)
	if err != nil {
		return
	}
	retObj, retCode, err = GetCmdClass(appid, classID)
}
