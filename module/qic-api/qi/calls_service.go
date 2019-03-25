package qi

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"
	"strconv"

	"emotibot.com/emotigo/pkg/logger"
	uuid "github.com/satori/go.uuid"

	"encoding/hex"
	"time"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

//ErrNotFound is indicated the resource is asked, but nowhere to found it.
var ErrNotFound = errors.New("resource not found")

// dao dependencies for call service.
var (
	callCount    = callDao.Count
	calls        = callDao.Calls
	valuesKey    = userValueDao.ValuesKey
	newUserValue = userValueDao.NewUserValue
	userValues   = userValueDao.UserValues
	userKeys     = userKeyDao.UserKeys
	keyvalues    = userKeyDao.KeyValues
)

func HasCall(id int64) (bool, error) {
	count, err := callDao.Count(nil, model.CallQuery{
		ID: []int64{id},
	})
	if err != nil {
		return false, fmt.Errorf("dao query failed, %v", err)
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

var (
	call = func(callUUID string, enterprise string) (c model.Call, err error) {
		query := model.CallQuery{
			UUID: []string{callUUID},
		}
		if enterprise != "" {
			query.EnterpriseID = &enterprise
		}
		calls, err := callDao.Calls(nil, query)
		if err != nil {
			return c, fmt.Errorf("get calls failed, %v", err)
		}
		if len(calls) < 1 {
			return c, ErrNotFound
		}
		return calls[0], nil
	}
	callRespsWithTotal = func(query model.CallQuery) (responses []CallResp, total int64, err error) {
		total, err = callCount(nil, query)
		if err != nil {
			return nil, 0, fmt.Errorf("call dao count query failed, %v", err)
		}
		calls, err := calls(nil, query)
		if err != nil {
			return nil, 0, fmt.Errorf("call dao call query failed, %v", err)
		}
		values, err := userValueDao.ValuesKey(nil, model.UserValueQuery{
			Type:     []int8{model.UserValueTypCall},
			ParentID: query.ID,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("find user value failed, %v", err)
		}

		tasks, err := TasksByCalls(calls)
		if err != nil {
			return nil, 0, fmt.Errorf("find tasks by calls failed, %v", err)
		}
		var result = make([]CallResp, 0, len(calls))
		for _, c := range calls {
			var t *model.Task
			for _, t = range tasks {
				if t.ID == c.TaskID {
					break
				}
			}
			var transaction int8
			if t.IsDeal {
				transaction = 1
			}
			callCustomCols := map[string]interface{}{}
			for _, v := range values {
				if v.LinkID == c.ID {
					switch v.UserKey.Type {
					case model.UserKeyTypArray:
						if rawdata, exist := callCustomCols[v.UserKey.InputName]; !exist {
							data := []string{v.Value}
							callCustomCols[v.UserKey.InputName] = data
						} else {
							data, ok := rawdata.([]string)
							if !ok {
								return nil, 0, fmt.Errorf("call %d value %d said it is array, but %s is not valid", c.ID, v.ID, v.Value)
							}
							data = append(data, v.Value)
						}

					case model.UserKeyTypNumber:
						intVal, err := strconv.Atoi(v.Value)
						if err == nil {
							callCustomCols[v.UserKey.InputName] = intVal
						} else {
							fltVal, err := strconv.ParseFloat(v.Value, 64)
							if err != nil {
								return nil, 0, fmt.Errorf("call %d value %d said it is number, but %s is not valid", c.ID, v.ID, v.Value)
							}
							callCustomCols[v.UserKey.InputName] = fltVal
						}
					case model.UserKeyTypTime:
						fallthrough
					case model.UserKeyTypString:
						fallthrough
					default:
						callCustomCols[v.UserKey.InputName] = v.Value
					}
				}
			}

			r := CallResp{
				CallID:        c.ID,
				CallUUID:      c.UUID,
				FileName:      *c.FileName,
				CallTime:      c.CallUnixTime,
				CallComment:   *c.Description,
				Transaction:   transaction,
				Series:        t.Series,
				HostID:        c.StaffID,
				HostName:      c.StaffName,
				Extension:     c.Ext,
				Department:    c.Department,
				CustomerID:    c.CustomerID,
				CustomerName:  c.CustomerName,
				CustomerPhone: c.CustomerPhone,
				LeftChannel:   callRoleTypStr(c.LeftChanRole),
				RightChannel:  callRoleTypStr(c.RightChanRole),
				Status:        int64(c.Status),
				UploadTime:    c.UploadUnixTime,
				CallLength:    float64(c.DurationMillSecond) / 1000,
				LeftSpeed:     c.LeftSpeed,
				RightSpeed:    c.RightSpeed,
				CustomColumns: callCustomCols,
			}
			if c.LeftSilenceTime != nil {
				r.LeftSilenceTime = *c.LeftSilenceTime
				r.LeftSilenceRate = (r.LeftSilenceTime * 1000.0) / float64(c.DurationMillSecond)
			}
			if c.RightSilenceTime != nil {
				r.RightSilenceRate = *c.RightSilenceTime
				r.RightSilenceRate = (r.RightSilenceTime * 1000.0) / float64(c.DurationMillSecond)
			}
			result = append(result, r)
		}

		return result, total, nil
	}
)

// Call return a call by given callID and enterprise
// If enterprise is empty, it will ignore it in conditions.
// If callID can not found, a ErrNotFound will returned
func Call(callUUID string, enterprise string) (c model.Call, err error) {
	return call(callUUID, enterprise)
}

// Calls is just a wrapper for callDao's calls.
// any usage need a more convenient service that retrive one element or with its task and user values.
// Should try Call or CallRespsWithTotal.
// It is consider to be deprecate or refractor.
func Calls(delegatee model.SqlLike, query model.CallQuery) ([]model.Call, error) {
	return calls(delegatee, query)
}

// CallRespsWithTotal query the call and  tasks and user values to assemble the call responses.
// It also returned the total count for call if query pagination is not nil.
func CallRespsWithTotal(query model.CallQuery) (responses []CallResp, total int64, err error) {
	return callRespsWithTotal(query)
}

//ErrCCTypeMismatch indicate the income call request has wrong data type of custom column.
var ErrCCTypeMismatch = errors.New("column type mismatch")

//NewCall create a call based on the input.
func NewCall(c *NewCallReq) (*model.Call, error) {
	var err error
	// create new call task
	tx, err := dbLike.Begin()
	if err != nil {
		return nil, fmt.Errorf("error while get transaction, %v", err)
	}
	defer tx.Rollback()
	timestamp := time.Now().Unix()
	var cTask *model.Task
	// call's task is determine by its serial number, if the same then use the old one.
	tasks, err := taskDao.Task(tx, model.TaskQuery{
		SN: []string{c.Series},
	})
	if err != nil {
		return nil, fmt.Errorf("fetch task failed, %v", err)
	}
	if len(tasks) >= 1 {
		cTask = &tasks[0]
		// TODO: Update time
		// cTask.UpdatedTime = timestamp
	} else {
		cTask, err = taskDao.NewTask(tx, model.Task{
			Status:      int8(0),
			Series:      c.Series,
			IsDeal:      c.Transaction == 1,
			CreatedTime: timestamp,
			UpdatedTime: timestamp,
		})
		if err != nil {
			return nil, fmt.Errorf("create new task failed, %v", err)
		}
	}

	// create uuid for call
	uid, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("new uuid failed, %v", err)
	}

	call := &model.Call{
		UUID:           hex.EncodeToString(uid[:]),
		FileName:       &c.FileName,
		Description:    &c.Description,
		UploadUnixTime: time.Now().Unix(),
		CallUnixTime:   c.CallTime,
		StaffID:        c.HostID,
		StaffName:      c.HostName,
		Ext:            c.Extension,
		Department:     c.Department,
		CustomerID:     c.GuestID,
		CustomerName:   c.GuestName,
		CustomerPhone:  c.GuestPhone,
		EnterpriseID:   c.Enterprise,
		UploadUser:     c.UploadUser,
		LeftChanRole:   callRoleTyp(c.LeftChannel),
		RightChanRole:  callRoleTyp(c.RightChannel),
		TaskID:         cTask.ID,
		Type:           c.Type,
	}

	calls, err := callDao.NewCalls(tx, []model.Call{*call})
	if err != nil {
		return nil, err
	}
	call = &calls[0]
	err = updateCallCustomInfo(tx, c, call, timestamp)
	if err != nil {
		return nil, err
	}

	return call, nil
}

func updateCallCustomInfo(tx model.SQLTx, c *NewCallReq, call *model.Call, timestamp int64) error {
	isEnable := true
	groups, err := serviceDAO.Group(tx, model.GroupQuery{
		EnterpriseID: c.Enterprise,
		IsEnable:     &isEnable,
	})
	if err != nil {
		return fmt.Errorf("query group failed, %v", err)
	}

	keys, err := userKeys(tx, model.UserKeyQuery{Enterprise: c.Enterprise})
	if err != nil {
		return fmt.Errorf("fetch key values failed, %v", err)
	}
	// filtered customcolumns from user
	var userInputs = make(map[string][]string)
	for name, value := range c.CustomColumns {
		var (
			k       model.UserKey
			isValid bool
		)
		for _, k = range keys {
			if name == k.InputName {
				isValid = true
				break
			}
		}
		if !isValid {
			continue
		}
		uv := make([]model.UserValue, 0)
		switch k.Type {
		case model.UserKeyTypArray:
			values, ok := value.([]interface{})
			if !ok {
				return ErrCCTypeMismatch
			}
			for _, val := range values {
				v := model.UserValue{
					LinkID:     call.ID,
					UserKeyID:  k.ID,
					Type:       model.UserValueTypCall,
					Value:      fmt.Sprintf("%s", val),
					CreateTime: timestamp,
					UpdateTime: timestamp,
				}
				uv = append(uv, v)
			}
		case model.UserKeyTypNumber:
			_, ok := value.(float64)
			if !ok {
				_, isInt := value.(int)
				if !isInt {
					return ErrCCTypeMismatch
				}
			}
			v := model.UserValue{
				LinkID:     call.ID,
				UserKeyID:  k.ID,
				Type:       model.UserValueTypCall,
				Value:      fmt.Sprintf("%s", value),
				CreateTime: timestamp,
				UpdateTime: timestamp,
			}
			uv = append(uv, v)
		case model.UserKeyTypTime:
			_, ok := value.(int)
			if !ok {
				return ErrCCTypeMismatch
			}
			v := model.UserValue{
				LinkID:     call.ID,
				UserKeyID:  k.ID,
				Type:       model.UserValueTypCall,
				Value:      fmt.Sprintf("%s", value),
				CreateTime: timestamp,
				UpdateTime: timestamp,
			}
			uv = append(uv, v)
		case model.UserKeyTypString:
			v := model.UserValue{
				LinkID:     call.ID,
				UserKeyID:  k.ID,
				Type:       model.UserValueTypCall,
				Value:      fmt.Sprintf("%s", value),
				CreateTime: timestamp,
				UpdateTime: timestamp,
			}
			uv = append(uv, v)
		}
		for _, v := range uv {
			v, err = newUserValue(tx, v)
			if err != nil {
				return fmt.Errorf("new user values failed, %v", err)
			}
			userInputs[name] = insertAndOrderStrings(userInputs[name], v.Value)
		}
	}

	//TODO: add matching for custom columns and conditions.
	valueQuery := model.UserValueQuery{
		Type: []int8{model.UserValueTypGroup},
	}
	for _, grp := range groups {
		valueQuery.ParentID = append(valueQuery.ParentID, grp.ID)
	}
	keyQuery := model.UserKeyQuery{
		Enterprise: c.Enterprise,
	}
	conditions, err := keyvalues(tx, keyQuery, valueQuery)
	if err != nil {
		return fmt.Errorf("fetch conditions failed, %v", err)
	}
	groupConditions := make(map[int64][]model.UserKey, len(groups))
	for _, cond := range conditions {
		for _, val := range cond.UserValues {
			c := *cond
			c.UserValues = []model.UserValue{val}
			hasKey := false
			keys := groupConditions[val.LinkID]
			for _, k := range keys {
				if k.ID == c.ID {
					k.UserValues = append(k.UserValues, val)
				}
			}
			if !hasKey {
				keys = append(keys, c)
			}
		}
	}
	matchedGroups := MatchGroup(matchDefaultConditions(groups, *call), groupConditions, userInputs)
	_, err = callDao.SetRuleGroupRelations(tx, *call, matchedGroups)
	if err != nil {
		return fmt.Errorf("set rule group failed, %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit new call transaction failed, %v", err)
	}

	return nil
}

//UpdateCall update the call data source
func UpdateCall(call *model.Call) error {
	return callDao.SetCall(nil, *call)
}

//ConfirmCall is the workflow to update call File Path and send the request into message queue.
func ConfirmCall(call *model.Call) error {
	type VAD struct {
		SegmentID int64   `json:"segment_id"`
		Speaker   int8    `json:"speaker"`
		StartTime float64 `json:"start_time"`
		EndTime   float64 `json:"end_time"`
	}

	type ASRInput struct {
		Version  float64 `json:"version"`
		CallID   string  `json:"call_id"`
		CallUUID string  `json:"call_uuid"`
		Path     string  `json:"path"`
		VADList  []*VAD  `json:"vad_list"`
	}

	//TODO: if call already Confirmed, it should not be able to
	if call.FilePath == nil {
		return fmt.Errorf("call FilePath should not be nil")
	}
	call.Status = model.CallStatusRunning
	err := UpdateCall(call)
	//TODO: ADD Task update too.
	if err != nil {
		return fmt.Errorf("update call db failed, %v", err)
	}
	// Because ASR expect us to give its real system filepath
	// which we only can hard coded or inject from env.
	// TODO: TELL ASR TEAM TO FIX IT!!!
	basePath, found := ModuleInfo.Environments["ASR_HARDCODE_VOLUME"]
	if !found {
		logger.Warn.Println("expect ASR_HARDCODE_VOLUME have setup, or asr may not be able to read the path.")
	}
	input := ASRInput{
		Version:  1.0,
		CallID:   strconv.FormatInt(call.ID, 10),
		CallUUID: call.UUID,
		Path:     path.Join(basePath, *call.FilePath),
	}

	if call.Type == model.CallTypeRealTime {
		vadList := []*VAD{}

		// Get call's segments to create VAD list of the call
		segments, err := getSegments(*call)
		if err != nil {
			return err
		}

		for _, segment := range segments {
			vad := VAD{
				SegmentID: segment.SegmentID,
				Speaker:   callRoleTyp(segment.Speaker),
				StartTime: segment.StartTime,
				EndTime:   segment.EndTime,
			}

			vadList = append(vadList, &vad)
		}

		input.VADList = vadList
	}

	resp, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("marshal asr input failed, %v", err)
	}
	err = producer.Produce(resp)
	if err != nil {
		return fmt.Errorf("publish failed, %v", err)
	}
	return nil
}

// MatchGroup filter the given groups by the groupConditions and userInputs.
// groupConditions
func MatchGroup(groups []model.Group, groupConditions map[int64][]model.UserKey, userInputs map[string][]string) []model.Group {
	var matchedGroups = make([]model.Group, 0)
	for _, grp := range groups {
		isValid := true
		conds := groupConditions[grp.ID]
		// group's conditions need to be all matched.
		if len(userInputs) < len(conds) {
			continue
		}
		for _, cond := range conds {
			uv, exist := userInputs[cond.InputName]
			if !exist {
				isValid = false
				break
			}
			//simple intersect n^2
			//TODO: change to sorted or hash
			var intersect = make([]string, 0)
			for _, kv := range cond.UserValues {
				for _, v := range uv {
					if kv.Value == v {
						intersect = append(intersect, v)
					}
				}
			}
			// need at least one value matched
			if len(intersect) == 0 {
				isValid = false
			}

		}
		if isValid {
			matchedGroups = append(matchedGroups, grp)
		}
	}
	return matchedGroups
}

func matchDefaultConditions(groups []model.Group, c model.Call) []model.Group {
	var matchedGroups = make([]model.Group, 0, len(groups))
	for _, grp := range groups {
		if grp.Condition == nil {
			continue
		}
		cond := grp.Condition
		if cond.Type == model.GroupCondTypOff {
			matchedGroups = append(matchedGroups, grp)
			continue
		}
		if cond.FileName != "" && cond.FileName != *c.FileName {
			continue
		}
		// if cond.Deal != -1 && cond.Deal != c.Transaction {
		// 	continue
		// }
		// if cond.Series != "" && cond.Series != c. {
		// 	continue
		// }
		if cond.UploadTimeStart != 0 && cond.UploadTimeStart > c.UploadUnixTime {
			continue
		}
		if cond.UploadTimeEnd != 0 && c.UploadUnixTime > cond.UploadTimeEnd {
			continue
		}
		if cond.StaffID != "" && c.StaffID != cond.StaffID {
			continue
		}
		if cond.StaffName != "" && cond.StaffName != c.StaffName {
			continue
		}
		if cond.Extension != "" && cond.Extension != c.Ext {
			continue
		}
		if cond.Department != "" && cond.Department != c.Department {
			continue
		}
		if cond.CustomerID != "" && cond.CustomerID != c.CustomerID {
			continue
		}
		if cond.CustomerName != "" && cond.CustomerName != c.CustomerName {
			continue
		}
		if cond.CustomerPhone != "" && cond.CustomerPhone != c.CustomerPhone {
			continue
		}
		if cond.CallStart != 0 && cond.CallStart > c.CallUnixTime {
			continue
		}
		if cond.CallEnd != 0 && c.CallUnixTime > cond.CallEnd {
			continue
		}
		if cond.LeftChannel != model.GroupCondRoleAny && cond.LeftChannel != c.LeftChanRole {
			continue
		}
		if cond.RightChannel != model.GroupCondRoleAny && cond.RightChannel != c.RightChanRole {
			continue
		}

		matchedGroups = append(matchedGroups, grp)
	}
	return matchedGroups
}

// insertAndOrderStrings insert v into orderedValues by ascending order.
func insertAndOrderStrings(orderedValues []string, v string) []string {
	values := make([]string, len(orderedValues)+1)
	index := sort.Search(len(orderedValues), func(i int) bool {
		return orderedValues[i] > v
	})
	// tail only need append
	if index == len(orderedValues) {
		values = append(orderedValues, v)
		return values
	}
	copy(values, orderedValues[:index])
	values[index] = v
	copy(values[index+1:], orderedValues[index:])
	return values
}
