package BF

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
)

func handleGetCmds(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	status := http.StatusOK
	retCode := ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()
	appid := util.GetAppID(r)

	cmds, err := GetCmds(appid)
	if err != nil {
		status = http.StatusInternalServerError
		retCode = ApiError.DB_ERROR
		retObj = err.Error()
	} else {
		retObj = cmds
	}
}
func handleGetCmd(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	status := http.StatusOK
	retCode := ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()
	appid := util.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		status = http.StatusBadRequest
		retCode = ApiError.REQUEST_ERROR
		retObj = err.Error()
		return
	}

	cmd, err := GetCmd(appid, id)
	if err != nil {
		status = http.StatusInternalServerError
		retCode = ApiError.DB_ERROR
		retObj = err.Error()
	} else if cmd == nil {
		status = http.StatusNotFound
		retCode = ApiError.REQUEST_ERROR
	} else {
		retObj = cmd
	}
}
func handleUpdateCmd(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	appid := util.GetAppID(r)
	status := http.StatusOK
	retCode := ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
			util.ConsulUpdateCmd(appid)
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()

	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		status, retCode = http.StatusBadRequest, ApiError.REQUEST_ERROR
		return
	}

	origCmd, err := GetCmd(appid, id)
	if err != nil {
		status, retCode = http.StatusInternalServerError, ApiError.IO_ERROR
		retObj = err.Error()
		return
	}
	if origCmd == nil {
		status, retCode = http.StatusNotFound, ApiError.REQUEST_ERROR
		return
	}

	cmd, err := parseCmdFromRequest(r)
	if err != nil {
		status, retCode = http.StatusBadRequest, ApiError.REQUEST_ERROR
		return
	}

	err = UpdateCmd(appid, id, cmd)
	if err != nil {
		status, retCode = http.StatusInternalServerError, ApiError.IO_ERROR
		retObj = err.Error()
		return
	}
	retObj = cmd
}
func handleAddCmd(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	appid := util.GetAppID(r)
	status := http.StatusOK
	retCode := ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
			util.ConsulUpdateCmd(appid)
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()

	cmd, err := parseCmdFromRequest(r)
	if err != nil {
		status = http.StatusBadRequest
		retCode = ApiError.REQUEST_ERROR
		retObj = err.Error()
		return
	}
	cid, err := strconv.Atoi(r.FormValue("cid"))
	if err != nil {
		retCode, retObj = ApiError.REQUEST_ERROR, fmt.Sprintf("get cid fail: %s", err.Error())
		return
	}

	id, err := AddCmd(appid, cmd, cid)
	if err != nil {
		status = http.StatusInternalServerError
		retCode = ApiError.IO_ERROR
		retObj = err.Error()
		return
	}
	cmd.ID = id
	retObj = cmd
}
func handleDeleteCmd(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	appid := util.GetAppID(r)
	status := http.StatusOK
	retCode := ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
			util.ConsulUpdateCmd(appid)
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		status = http.StatusBadRequest
		retCode = ApiError.REQUEST_ERROR
		retObj = err.Error()
		return
	}

	err = DeleteCmd(appid, id)
	if err != nil {
		status = http.StatusInternalServerError
		retCode = ApiError.DB_ERROR
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
	ret.Cmd = ruleContents

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
	status, retCode := http.StatusOK, ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()
	appid := util.GetAppID(r)
	cmdID, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		status, retCode = http.StatusBadRequest, ApiError.REQUEST_ERROR
		return
	}

	labels, err := GetLabelsOfCmd(appid, cmdID)
	if err != nil {
		status, retCode = http.StatusInternalServerError, ApiError.DB_ERROR
		retObj = err.Error()
	} else {
		retObj = labels
	}
}
