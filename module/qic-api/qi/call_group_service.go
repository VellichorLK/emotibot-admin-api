package qi

import (
	"encoding/json"
	"time"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	callGroupDao       model.CallGroupDao       = &model.CallGroupSQLDao{}
	creditCallGroupDao model.CreditCallGroupDao = &model.CreditCallGroupSQLDao{}
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
			callGroupID, callIDs, err := groupCalls(call.EnterpriseID, callResp, cgCond)
			if err != nil {
				logger.Error.Printf("group calls failed. call.ID: %d, error: %s\n", call.ID, err)
				return err
			}
			creditTree, ruleIDs, err := GetCallGroupCreditTree(callIDs)
			if err != nil {
				logger.Error.Printf("get call group credit tree failed. error: %s\n", err)
				return err
			}
			_, err = CreateCreditCallGroups(uint64(callGroupID), creditTree, ruleIDs)
			if err != nil {
				logger.Error.Printf("create call group credit failed. error: %s\n", err)
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

func groupCalls(enterpriseID string, callResp CallResp, cgCond *model.CGCondition) (int64, []int64, error) {
	if dbLike == nil {
		return 0, nil, ErrNilCon
	}
	tx, err := dbLike.Begin()
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback()

	uploadTime := time.Unix(callResp.UploadTime, 0)
	uploadTimeUnix := uploadTime.Unix()
	callTime := time.Unix(callResp.CallTime, 0)
	callTimeUnix := callTime.Unix()

	// get call list of those which have the targe groupBy UserValue
	groupBy := cgCond.GroupBy[0] // TODO: implement multiple groupBy keys
	values := callResp.CustomColumns[groupBy.Inputname].([]string)
	valueType := model.UserValueTypCall
	fromTime := callTime.AddDate(0, 0, cgCond.DayRange*-1).Unix()
	toTime := callTimeUnix
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
		return 0, nil, err
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
		return 0, nil, err
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
		return 0, nil, nil
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
		return 0, nil, err
	}

	// create a new call group
	callGroup := model.CallGroup{
		IsDelete:             0,
		CallGroupConditionID: cgCond.ID,
		Enterprise:           enterpriseID,
		LastCallID:           callResp.CallID,
		LastCallTime:         callTimeUnix,
		CreateTime:           uploadTimeUnix,
		UpdateTime:           uploadTimeUnix,
		Calls:                callsToGroup,
	}
	callGroupID, err := callGroupDao.CreateCallGroup(tx, &callGroup)
	if err != nil {
		logger.Error.Printf("failed to create CallGroup. %s\n", err)
		return 0, nil, err
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
			return 0, nil, err
		}
	}
	return callGroupID, callGroup.Calls, tx.Commit()
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
		isGroup := bool(false)
		if len(callGroup.Calls) > 1 {
			isGroup = true
		}
		resp := GroupedCallsResp{
			Setting: callRespMap[callGroup.LastCallID],
			IsGroup: isGroup,
			Calls:   callResps,
		}
		respList = append(respList, &resp)
	}

	for id, callResp := range callRespMap {
		if _, ok := callRespsInGroupMap[id]; !ok {
			callResps := []*CallResp{}
			resp := GroupedCallsResp{
				Setting: callResp,
				IsGroup: false,
				Calls:   callResps,
			}
			respList = append(respList, &resp)
		}
	}

	// out, _ = json.Marshal(respList)
	// logger.Trace.Printf("respList")
	// logger.Trace.Printf(string(out))
	return respList, total, nil
}

// CallGroupCreditTree defines the call group credit tree structure
type CallGroupCreditTree struct {
	Credit       *model.SimpleCredit
	RuleGroupMap map[uint64]*ruleGroupCredit
}
type ruleGroupCredit struct {
	Credit  *model.SimpleCredit
	RuleMap map[uint64]*ruleCredit
	Rules   map[uint64]*[]*model.SimpleCredit
}
type ruleCredit struct {
	Credit   *model.SimpleCredit
	CFlowMap map[uint64]*convFlowCredit
}
type convFlowCredit struct {
	Credit      *model.SimpleCredit
	SenGroupMap map[uint64]*senGroupCredit
}
type senGroupCredit struct {
	Credit *model.SimpleCredit
	SenMap map[uint64]*senCredit
}
type senCredit struct {
	Credit *model.SimpleCredit
	SegMap map[uint64]*segCredit
}
type segCredit struct {
	Credit *model.SimpleCredit
}

// GetCallGroupCreditTree return the requested CallGroupCreditTree tree and the ID list of Rules
func GetCallGroupCreditTree(inputCallIDs []int64) (*CallGroupCreditTree, []uint64, error) {
	// inputCallIDs = []int64{153}
	out, _ := json.Marshal(inputCallIDs)
	logger.Trace.Printf("inputCallIDs")
	logger.Trace.Printf(string(out))
	if dbLike == nil {
		return nil, nil, ErrNilCon
	}
	tx, err := dbLike.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	callIDs := make([]uint64, len(inputCallIDs))
	for i := 0; i < len(inputCallIDs); i++ {
		callIDs[i] = uint64(inputCallIDs[i])
	}
	creditQuery := model.CreditQuery{
		Calls: callIDs,
	}
	credits, err := creditDao.GetCallCredit(tx, &creditQuery)
	if err != nil {
		logger.Error.Printf("get credits failed\n")
		return nil, nil, err
	}

	var ruleIDs []uint64
	rgOrgIDMap := make(map[uint64]uint64)
	rOrgIDMap := make(map[uint64]uint64)
	cfOrgIDMap := make(map[uint64]uint64)
	senGrpOrgIDMap := make(map[uint64]uint64)
	senOrgIDMap := make(map[uint64]uint64)
	segOrgIDMap := make(map[uint64]uint64)

	rgCreditMap := make(map[uint64]*ruleGroupCredit)
	rCreditMap := make(map[uint64]*ruleCredit)
	cfCreditMap := make(map[uint64]*convFlowCredit)
	senGrpCreditMap := make(map[uint64]*senGroupCredit)
	senCreditMap := make(map[uint64]*senCredit)
	segCreditMap := make(map[uint64]*segCredit)

	cgCredit := CallGroupCreditTree{
		Credit:       &model.SimpleCredit{},
		RuleGroupMap: make(map[uint64]*ruleGroupCredit),
	}
	for _, credit := range credits {
		switch levelType(credit.Type) {
		// case levCallType:
		// 	continue
		case levRuleGrpTyp:
			rgOrgIDMap[credit.ID] = credit.OrgID
			if _, ok := rgCreditMap[credit.OrgID]; !ok {
				rgCredit := ruleGroupCredit{
					Credit:  credit,
					RuleMap: make(map[uint64]*ruleCredit),
					Rules:   make(map[uint64]*[]*model.SimpleCredit),
				}
				rgCreditMap[credit.OrgID] = &rgCredit
			}
		// case levRuleTyp, levSilenceTyp, levSpeedTyp, levInterposalTyp:
		case levRuleTyp:
			rOrgIDMap[credit.ID] = credit.OrgID
			rCredit, ok := rCreditMap[credit.OrgID]
			if !ok {
				newCredit := ruleCredit{
					Credit:   credit,
					CFlowMap: make(map[uint64]*convFlowCredit),
				}
				rCreditMap[credit.OrgID] = &newCredit
				rCredit = &newCredit
			}

			parentOrgID := rgOrgIDMap[credit.ParentID]
			parentCredit := rgCreditMap[parentOrgID]
			if _, ok := parentCredit.RuleMap[credit.OrgID]; !ok {
				parentCredit.RuleMap[credit.OrgID] = rCredit
			}

			rgCredit := parentCredit
			rOrgID := credit.OrgID
			if rgCredit.Rules[rOrgID] == nil {
				creditList := []*model.SimpleCredit{credit}
				rgCredit.Rules[rOrgID] = &creditList
			} else {
				*rgCredit.Rules[rOrgID] = append(*rgCredit.Rules[rOrgID], credit)
			}
			ruleIDs = append(ruleIDs, credit.OrgID)
		case levCFTyp:
			cfOrgIDMap[credit.ID] = credit.OrgID
			cfCredit, ok := cfCreditMap[credit.OrgID]
			if !ok {
				newCredit := convFlowCredit{
					Credit:      credit,
					SenGroupMap: make(map[uint64]*senGroupCredit),
				}
				cfCreditMap[credit.OrgID] = &newCredit
				cfCredit = &newCredit
			}
			parentOrgID := rOrgIDMap[credit.ParentID]
			parentCredit := rCreditMap[parentOrgID]
			if _, ok := parentCredit.CFlowMap[credit.OrgID]; !ok {
				parentCredit.CFlowMap[credit.OrgID] = cfCredit
			}
		case levSenGrpTyp:
			senGrpOrgIDMap[credit.ID] = credit.OrgID
			sgCredit, ok := senGrpCreditMap[credit.OrgID]
			if !ok {
				newCredit := senGroupCredit{
					Credit: credit,
					SenMap: make(map[uint64]*senCredit),
				}
				senGrpCreditMap[credit.OrgID] = &newCredit
				sgCredit = &newCredit
			}
			parentOrgID := cfOrgIDMap[credit.ParentID]
			parentCredit := cfCreditMap[parentOrgID]
			if _, ok := parentCredit.SenGroupMap[credit.OrgID]; !ok {
				parentCredit.SenGroupMap[credit.OrgID] = sgCredit
			}
		case levSenTyp:
			senOrgIDMap[credit.ID] = credit.OrgID
			sCredit, ok := senCreditMap[credit.OrgID]
			if !ok {
				newCredit := senCredit{
					Credit: credit,
					SegMap: make(map[uint64]*segCredit),
				}
				senCreditMap[credit.OrgID] = &newCredit
				sCredit = &newCredit
			}
			parentOrgID := senGrpOrgIDMap[credit.ParentID]
			parentCredit := senGrpCreditMap[parentOrgID]
			if _, ok := parentCredit.SenMap[credit.OrgID]; !ok {
				parentCredit.SenMap[credit.OrgID] = sCredit
			}
		case levSegTyp:
			segOrgIDMap[credit.ID] = credit.OrgID
			sCredit, ok := segCreditMap[credit.OrgID]
			if !ok {
				newCredit := segCredit{
					Credit: credit,
				}
				segCreditMap[credit.OrgID] = &newCredit
				sCredit = &newCredit
			}
			parentOrgID := senOrgIDMap[credit.ParentID]
			parentCredit := senCreditMap[parentOrgID]
			if _, ok := parentCredit.SegMap[credit.OrgID]; !ok {
				parentCredit.SegMap[credit.OrgID] = sCredit
			}
		default:
			//logger.Error.Printf("credit result %d id has the unknown type %d\n", v.ID, v.Type)
			continue
		}
	}

	// out, _ = json.Marshal(credits)
	// logger.Trace.Printf("credits")
	// logger.Trace.Printf(string(out))
	cgCredit.RuleGroupMap = rgCreditMap
	out, _ = json.Marshal(cgCredit)
	logger.Trace.Printf("cgCredit")
	logger.Trace.Printf(string(out))
	return &cgCredit, ruleIDs, tx.Commit()
}

// CallGroupCreditCGTree defines the call group CreditCallGroup tree structure
type CallGroupCreditCGTree struct {
	Credit       *model.CreditCallGroup
	RuleGroupMap map[uint64]*ruleGroupCreditCG
}
type ruleGroupCreditCG struct {
	Credit  *model.CreditCallGroup
	RuleMap map[uint64]*ruleCreditCG
	Rules   map[uint64]*[]*model.SimpleCredit
}
type ruleCreditCG struct {
	Credit   *model.CreditCallGroup
	CFlowMap map[uint64]*convFlowCreditCG
}
type convFlowCreditCG struct {
	Credit      *model.CreditCallGroup
	SenGroupMap map[uint64]*senGroupCreditCG
}
type senGroupCreditCG struct {
	Credit *model.CreditCallGroup
	SenMap map[uint64]*senCreditCG
}
type senCreditCG struct {
	Credit *model.CreditCallGroup
	SegMap map[uint64]*segCreditCG
}
type segCreditCG struct {
	Credit *model.CreditCallGroup
}

// CreateCreditCallGroups calculate the call group scores and store it
func CreateCreditCallGroups(callGroupID uint64, creditTree *CallGroupCreditTree, ruleIDs []uint64) (*CallGroupCreditCGTree, error) {
	out, _ := json.Marshal(ruleIDs)
	logger.Trace.Printf("ruleIDs")
	logger.Trace.Printf(string(out))
	if dbLike == nil {
		return nil, ErrNilCon
	}
	tx, err := dbLike.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ruleList, err := conversationRuleDao.GetBy(&model.ConversationRuleFilter{ID: ruleIDs, Severity: -1, IsDeleted: -1}, tx)
	if err != nil {
		logger.Error.Printf("get rule list %+v failed. %s\n", ruleIDs, err)
		return nil, err
	}
	ruleMap := make(map[uint64]model.ConversationRule)
	for _, rule := range ruleList {
		ruleMap[uint64(rule.ID)] = rule
	}

	const (
		Valid   int = 1
		InValid int = 0
	)

	callGroupScore := int(100)
	createTime := time.Now().Unix()
	creditCG := &model.CreditCallGroup{
		CallGroupID: callGroupID, Type: 0, ParentID: 0, OrgID: 0, Valid: 0,
		Revise: 0, Score: callGroupScore, CreateTime: createTime, UpdateTime: createTime, CallID: 0,
	}
	creditCGTree := &CallGroupCreditCGTree{
		Credit:       creditCG,
		RuleGroupMap: make(map[uint64]*ruleGroupCreditCG),
	}
	parentCG, err := creditCallGroupDao.CreateCreditCallGroup(tx, creditCG)
	if err != nil {
		logger.Error.Printf("create call group credit %+v failed. %s\n", creditCG, err)
		return nil, err
	}

	for rgID, rgCredit := range creditTree.RuleGroupMap {
		rgScore := int(0)
		credit := rgCredit.Credit
		creditCG = &model.CreditCallGroup{
			CallGroupID: callGroupID, Type: credit.Type, ParentID: uint64(parentCG), OrgID: credit.OrgID, Valid: -1,
			Revise: -1, Score: rgScore, CreateTime: createTime, UpdateTime: createTime, CallID: 0,
		}
		rgCreditCG := &ruleGroupCreditCG{
			Credit:  creditCG,
			RuleMap: make(map[uint64]*ruleCreditCG),
			Rules:   make(map[uint64]*[]*model.SimpleCredit),
		}
		creditCGTree.RuleGroupMap[rgID] = rgCreditCG

		parentRG, err := creditCallGroupDao.CreateCreditCallGroup(tx, creditCG)
		if err != nil {
			logger.Error.Printf("create rule group credit %+v failed. %s\n", creditCG, err)
			return nil, err
		}

		for rID, rules := range rgCredit.Rules {
			convRule := ruleMap[rID]
			var valid int
			var callID uint64
			if convRule.Method == model.RuleMethodPositive {
				valid = InValid
				for _, rule := range *rules {
					if rule.Valid == Valid {
						valid = Valid
						callID = rule.CallID
						break
					}
				}
			} else if convRule.Method == model.RuleMethodNegative {
				valid = Valid
				for _, rule := range *rules {
					if rule.Valid == InValid {
						valid = InValid
						callID = rule.CallID
						break
					}
				}
			}
			score := int(0)
			isPosScore := convRule.Score > 0
			if valid == Valid && isPosScore {
				score = convRule.Score
			} else if valid == InValid && !isPosScore {
				score = convRule.Score
			}
			rgScore += score
			credit = rgCredit.RuleMap[rID].Credit
			creditCG = &model.CreditCallGroup{
				CallGroupID: callGroupID, Type: credit.Type, ParentID: uint64(parentRG), OrgID: credit.OrgID, Valid: valid,
				Revise: -1, Score: score, CreateTime: createTime, UpdateTime: createTime, CallID: callID,
			}
			rCreditCG := &ruleCreditCG{
				Credit:   creditCG,
				CFlowMap: make(map[uint64]*convFlowCreditCG),
			}
			rgCreditCG.RuleMap[rID] = rCreditCG
			_, err := creditCallGroupDao.CreateCreditCallGroup(tx, creditCG)
			if err != nil {
				logger.Error.Printf("create rule credit %+v failed. %s\n", creditCG, err)
				return nil, err
			}
		}
		rgCreditCG.Credit.Score = rgScore
		callGroupScore += rgScore
		updateSet := model.CreditCallGroupUpdateSet{Score: &rgScore}
		_, err = creditCallGroupDao.UpdateCreditCallGroup(tx, &model.GeneralQuery{ID: []int64{parentRG}}, &updateSet)
		if err != nil {
			logger.Error.Printf("update rule group credit %+v failed. %s\n", updateSet, err)
			return nil, err
		}
	}
	creditCGTree.Credit.Score = callGroupScore
	updateSet := model.CreditCallGroupUpdateSet{Score: &callGroupScore}
	_, err = creditCallGroupDao.UpdateCreditCallGroup(tx, &model.GeneralQuery{ID: []int64{parentCG}}, &updateSet)
	if err != nil {
		logger.Error.Printf("update call group credit %+v failed. %s\n", updateSet, err)
		return nil, err
	}
	return creditCGTree, nil
}
