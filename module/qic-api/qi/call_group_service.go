package qi

import (
	"time"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	callGroupDao model.CallGroupDao = &model.CallGroupSQLDao{}
)

// CreateCallGroupCondition create a new call group condition
func CreateCallGroupCondition(reqModel *model.CallGroupCondition, enterprise string) (int64, error) {
	if reqModel == nil {
		return 0, ErrNoArgument
	}
	if dbLike == nil {
		return 0, ErrNilCon
	}
	reqModel.Enterprise = enterprise
	reqModel.CreateTime = time.Now().Unix()
	reqModel.UpdateTime = reqModel.CreateTime
	id, err := callGroupDao.CreateCondition(dbLike.Conn(), reqModel)
	return id, err
}

// GetCallGroupConditionList return the detail of a call group condition
func GetCallGroupConditionList(query *model.GeneralQuery, pagination *model.Pagination) ([]*model.CallGroupCondition, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}
	return callGroupDao.GetConditionList(dbLike.Conn(), query, pagination)
}

//CountCallGroupCondition counts the total number of call group condition
func CountCallGroupCondition(query *model.GeneralQuery) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	return callGroupDao.CountCondition(dbLike.Conn(), query)
}

// UpdateCallGroupCondition update the call group condition
func UpdateCallGroupCondition(query *model.GeneralQuery, data *model.CallGroupConditionUpdateSet) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	if query == nil || len(query.ID) == 0 {
		return 0, ErrNoID
	}
	tx, err := dbLike.Begin()
	if err != nil {
		logger.Error.Printf("create session failed. %s\n", err)
		return 0, err
	}
	defer tx.Rollback()

	id, err := callGroupDao.UpdateCondition(tx, query, data)
	if err != nil {
		logger.Error.Printf("update failed. %s\n", err)
		return 0, err
	}
	tx.Commit()
	return id, nil

}

// DeleteCallGroupCondition delete the call group condition
func DeleteCallGroupCondition(query *model.GeneralQuery) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	if query == nil || len(query.ID) == 0 {
		return 0, ErrNoID
	}
	return callGroupDao.SoftDeleteCondition(dbLike.Conn(), query)
}

// CreateCallGroups create call groups
func CreateCallGroups(enterprise string, conditionID int64, data *model.CallGroupCreateList) error {
	return nil
}

// GroupCalls groups calls
func GroupCalls(call *model.Call) error {
	if dbLike == nil {
		return ErrNilCon
	}
	enterpriseID := "global"
	var zero, one int = 0, 1
	query := model.CGConditionQuery{
		Enterprise:      &enterpriseID,
		IsEnable:        &one,
		IsDelete:        &zero,
		UserKeyIsDelete: &zero,
	}
	cgConds, err := callGroupDao.GetCGCondition(dbLike.Conn(), &query)
	if err != nil {
		logger.Error.Printf("get call group condition failed. %s\n", err)
		return err
	}

	callResps, _, err := CallRespsWithTotal(model.CallQuery{
		ID: []int64{call.ID},
	})
	if err != nil {
		logger.Error.Printf("get call response with total failed. %s\n", err)
		return err
	}
	if len(callResps) == 0 {
		logger.Error.Printf("failed to find call with id: %d\n", call.ID)
	}
	callResp := callResps[0]
	for _, cgCond := range cgConds {
		var valid = true
		for inputname, validValue := range cgCond.FilterBy {
			vs, ok := callResp.CustomColumns[inputname]
			if !ok {
				valid = false
				break
			}
			values := vs.([]string)
			exist := contains(values, validValue)
			if !exist {
				valid = false
				break
			}
		}
		if valid {
			err = groupCalls(call.EnterpriseID, callResp, cgCond)
			if err != nil {
				logger.Error.Printf("group calls failed. call.ID: %d, error: %s\n", call.ID, err)
				return err
			}
		}
	}
	return nil
}

func contains(strings []string, s string) bool {
	for _, string := range strings {
		if string == s {
			return true
		}
	}
	return false
}

func groupCalls(enterpriseID string, callResp CallResp, cgCond *model.CGCondition) error {
	if dbLike == nil {
		return ErrNilCon
	}
	tx, err := dbLike.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	uploadTime := time.Unix(callResp.UploadTime, 0)
	uploadTimeUnix := uploadTime.Unix()

	// get call list of those which have the targe groupBy UserValue
	groupBy := cgCond.GroupBy[0] // TODO: implement multiple groupBy keys
	values := callResp.CustomColumns[groupBy.Inputname].([]string)
	valueType := model.UserValueTypCall
	fromTime := uploadTime.AddDate(0, 0, cgCond.DayRange*-1).Unix()
	toTime := uploadTimeUnix
	query := model.CallsToGroupQuery{
		UserValueType: &valueType,
		UserKeyID:     &groupBy.ID,
		UserValue:     &values,
		StartTime:     &fromTime,
		EndTime:       &toTime,
	}
	callIDs, err := callGroupDao.GetCallIDsToGroup(tx, &query)
	if err != nil {
		logger.Error.Printf("failed to get call ids to group . %s\n", err)
		return err
	}

	// get CallGroup list related to target callIDs
	isDelete := int(0)
	cgQuery := model.CallGroupQuery{
		IsDelete:  &isDelete,
		CallIDs:   &callIDs,
		StartTime: &fromTime,
		EndTime:   &toTime,
	}
	callGroupList, err := callGroupDao.GetCallGroups(tx, &cgQuery)
	if err != nil {
		logger.Error.Printf("failed to get CallGroup list. %s\n", err)
		return err
	}
	callGroupToDelete := []int64{}
	callsToGroup := []int64{}
	callMap := make(map[int64]bool)
	for _, callGroup := range callGroupList {
		callGroupToDelete = append(callGroupToDelete, callGroup.ID)
		for _, callID := range callGroup.Calls {
			if _, ok := callMap[callID]; !ok {
				callMap[callID] = true
				callsToGroup = append(callsToGroup, callID)
			}
		}
	}
	for _, callID := range callIDs {
		if _, ok := callMap[callID]; !ok {
			callMap[callID] = true
			callsToGroup = append(callsToGroup, callID)
		}
	}
	if len(callsToGroup) == 0 {
		return nil
	}

	// out, _ = json.Marshal(callGroupToDelete)
	// logger.Trace.Printf("callGroupToDelete")
	// logger.Trace.Printf(string(out))
	// out, _ = json.Marshal(callsToGroup)
	// logger.Trace.Printf("callsToGroup")
	// logger.Trace.Printf(string(out))

	// delete old CallGroups
	gQuery := model.GeneralQuery{
		ID: callGroupToDelete,
	}
	err = callGroupDao.SoftDeleteCallGroup(tx, &gQuery)
	if err != nil {
		logger.Error.Printf("failed to soft delete CallGroup list. %s\n", err)
		return err
	}

	// create a new call group
	callGroup := model.CallGroup{
		IsDelete:             0,
		CallGroupConditionID: cgCond.ID,
		Enterprise:           enterpriseID,
		CreateTime:           uploadTimeUnix,
		UpdateTime:           uploadTimeUnix,
		Calls:                callsToGroup,
	}
	callGroupID, err := callGroupDao.CreateCallGroup(tx, &callGroup)
	if err != nil {
		logger.Error.Printf("failed to create CallGroup. %s\n", err)
		return err
	}

	// create new relations between CallGroup and Call
	for _, callID := range callGroup.Calls {
		callGroupRelation := model.CallGroupRelation{
			CallGroupID: callGroupID,
			CallID:      callID,
		}
		_, err = callGroupDao.CreateCallGroupRelation(tx, &callGroupRelation)
		if err != nil {
			logger.Error.Printf("failed to create CallGroupRelation. %s\n", err)
			return err
		}
	}
	return tx.Commit()
}

// GroupedCallsResp defines the response structure of GetGroupedCalls
type GroupedCallsResp struct {
	Setting *CallResp   `json:"setting"`
	IsGroup bool        `json:"is_group"`
	Calls   []*CallResp `json:"calls"`
}

// GetGroupedCalls return the list of Calls and CallGroups
func GetGroupedCalls(query *model.CallQuery) ([]*GroupedCallsResp, int64, error) {
	total, err := callCount(nil, *query)
	if err != nil {
		logger.Error.Printf("failed to count calls. %s\n", err)
		return nil, 0, err
	}
	calls, err := calls(nil, *query)
	if err != nil {
		logger.Error.Printf("failed to get calls. %s\n", err)
		return nil, 0, err
	}
	if len(calls) == 0 {
		return []*GroupedCallsResp{}, 0, nil
	}

	callIDs := []int64{}
	callIDMap := make(map[int64]bool)
	mostRecentCallTime := int64(0)
	for _, call := range calls {
		callIDs = append(callIDs, call.ID)
		callIDMap[call.ID] = true
		if call.CallUnixTime > mostRecentCallTime {
			mostRecentCallTime = call.CallUnixTime
		}
	}

	// get CallGroups related to target callIDs
	isDelete := int(0)
	cgQuery := model.CallGroupQuery{
		IsDelete: &isDelete,
		CallIDs:  &callIDs,
	}
	callGroupList, err := callGroupDao.GetCallGroups(dbLike.Conn(), &cgQuery)
	if err != nil {
		logger.Error.Printf("failed to get CallGroup list. %s\n", err)
		return nil, 0, err
	}
	// append all calls in callGroupList to callIDs
	for _, callGroup := range callGroupList {
		for _, id := range callGroup.Calls {
			if _, ok := callIDMap[id]; !ok {
				callIDs = append(callIDs, id)
				callIDMap[id] = true
			}
		}
	}

	callQuery := model.CallQuery{
		ID: callIDs,
	}
	callResps, _, err := CallRespsWithTotal(callQuery)
	callRespMap := make(map[int64]*CallResp)
	for i := 0; i < len(callResps); i++ {
		callResp := callResps[i]
		callRespMap[callResp.CallID] = &callResp
	}

	respList := []*GroupedCallsResp{}
	callRespsInGroupMap := make(map[int64]bool)
	for _, callGroup := range callGroupList {
		callResps := []*CallResp{}
		for _, id := range callGroup.Calls {
			callResps = append(callResps, callRespMap[id])
			callRespsInGroupMap[id] = true
		}
		resp := GroupedCallsResp{
			IsGroup: true,
			Calls:   callResps,
		}
		respList = append(respList, &resp)
	}

	for id, callResp := range callRespMap {
		if _, ok := callRespsInGroupMap[id]; !ok {
			resp := GroupedCallsResp{
				Setting: callResp,
				IsGroup: false,
			}
			respList = append(respList, &resp)
		}
	}

	// out, _ = json.Marshal(respList)
	// logger.Trace.Printf("respList")
	// logger.Trace.Printf(string(out))
	return respList, total, nil
}
