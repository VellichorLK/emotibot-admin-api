package qi

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
)

var (
	navDao         model.NavigationDao  = &model.NavigationSQLDao{}
	navOnTheFlyDao model.NavOnTheFlyDao = &model.NavOnTheFlySQLDao{}
)

//NewFlow creates the new flow and sets the empty node and intent
func NewFlow(r *reqNewFlow, enterprise string) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	if r == nil {
		return 0, ErrNilFlow
	}
	uuid, err := general.UUID()
	if err != nil {
		logger.Error.Printf("generate uuid failed. %s\n", err)
		return 0, err
	}
	now := time.Now().Unix()
	q := &model.NavFlow{Name: r.Name, Enterprise: enterprise,
		CreateTime: now, UpdateTime: now,
		UUID: uuid}

	switch r.Type {
	case "intent":
		q.IntentName = r.IntentName
	default:
		q.IgnoreIntent = 1
	}

	id, err := navDao.NewFlow(dbLike.Conn(), q)
	if err != nil {
		logger.Error.Printf("Create new flow failed. %s\n", err)
	}
	return id, err
}

//NewNode generates a new sentence group and link to the navigation
func NewNode(nav int64, senGrp *model.SentenceGroup) error {
	if dbLike == nil {
		return ErrNilCon
	}
	if senGrp == nil {
		return ErrNilSentenceGroup
	}
	isDelete := 0
	q := &model.NavQuery{ID: []int64{nav}, IsDelete: &isDelete, Enterprise: &senGrp.Enterprise}
	flows, err := navDao.GetFlows(dbLike.Conn(), q, nil)
	if err != nil {
		logger.Error.Printf("get flow failed. %s\n", err)
		return err
	}
	if len(flows) == 0 {
		return ErrNilFlow
	}

	var nodeOrder []string
	if flows[0].NodeOrder != "" {
		err = json.Unmarshal([]byte(flows[0].NodeOrder), &nodeOrder)
		if err != nil {
			return fmt.Errorf("unmarshal node_order failed. %d. %s", flows[0].ID, err)
		}
	}

	tx, err := dbLike.Begin()
	if err != nil {
		logger.Error.Printf("get transaction failed. %s\n", err)
		return err
	}
	defer tx.Rollback()

	//it's not optional
	if senGrp.Optional == 0 {
		simpleSentences, err := simpleSentencesOf(senGrp, tx)
		if err != nil {
			logger.Error.Printf("get the sentence failed. %s\n", err)
			return err
		}
		senGrp.Sentences = simpleSentences
	}
	uuid, err := general.UUID()
	if err != nil {
		logger.Error.Printf("generate uuid failed. %s\n", err)
		return err
	}
	senGrp.UUID = uuid

	now := time.Now().Unix()
	senGrp.CreateTime = now
	senGrp.UpdateTime = now

	createGrp, err := sentenceGroupDao.Create(senGrp, tx)
	if err != nil {
		logger.Error.Printf("create sentence group failed. %s\n", err)
		return err
	}

	_, err = navDao.InsertRelation(tx, nav, createGrp.ID)
	if err != nil {
		logger.Error.Printf("insert nav to sentence group relation failed. %s\n", err)
		return err
	}

	nodeOrder = append(nodeOrder, uuid)
	err = updateNodeOrder(tx, nodeOrder, flows[0].ID)
	if err != nil {
		logger.Error.Printf("update node order failed. %d,%+v,  %s\n", flows[0].ID, nodeOrder, err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		logger.Error.Printf("commit data failed. %s\n", err)
	}
	return err
}

//UpdateNodeOrder adjusts the order of the nodes
func UpdateNodeOrder(nav int64, order []string) error {
	if dbLike == nil {
		return ErrNilCon
	}
	if len(order) == 0 {
		return nil
	}
	return updateNodeOrder(dbLike.Conn(), order, nav)

}

func updateNodeOrder(conn model.SqlLike, order []string, nav int64) error {
	orderStr, err := json.Marshal(order)
	if err != nil {
		logger.Error.Printf("marshal node order failed. %+v, %s\n", order, err)
	} else {
		_, err := navDao.UpdateNodeOrders(conn, nav, string(orderStr))
		if err != nil {
			logger.Error.Printf("update node order failed. %+v, %s\n", order, err)
		}
	}
	return err
}

//UpdateFlowName updates the flow, currently only the flow name
func UpdateFlowName(nav int64, enterprise string, name string) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}

	q := &model.NavQuery{ID: []int64{nav}, Enterprise: &enterprise}
	d := &model.NavFlowUpdate{Name: &name}
	affected, err := navDao.UpdateFlows(dbLike.Conn(), q, d)
	if err != nil {
		logger.Error.Printf("call update flow dao failed.%s\n", err)
	}
	return affected, err
}

//UpdateFlow updates the flow information in the database
func UpdateFlow(nav int64, enterprise string, d *model.NavFlowUpdate) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	if d == nil {
		return 0, ErrNilFlow
	}

	q := &model.NavQuery{ID: []int64{nav}, Enterprise: &enterprise}
	affected, err := navDao.UpdateFlows(dbLike.Conn(), q, d)
	if err != nil {
		logger.Error.Printf("call update flow dao failed.%s\n", err)
	}
	return affected, err
}

//DeleteFlow deletes the flow
func DeleteFlow(nav int64, enterprise string) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	tx, err := dbLike.Begin()
	if err != nil {
		logger.Error.Printf("get transaction failed. %s\n", err)
		return 0, err
	}
	defer tx.Rollback()
	q := &model.NavQuery{ID: []int64{nav}, Enterprise: &enterprise}
	affected, err := navDao.SoftDeleteFlows(tx, q)
	if err != nil {
		logger.Error.Printf("delete flow failed. %s\n", err)
		return 0, err
	}
	/*
		if affected != 0 {
			_, err = navDao.DeleteRelation(tx, nav)
			if err != nil {
				logger.Error.Printf("delete flow relation failed. %s\n", err)
			}
		}
	*/
	err = tx.Commit()
	return affected, err
}

//DetailNavFlow is the nav information and its relative setting
type DetailNavFlow struct {
	model.NavFlow
	model.SentenceGroup //the call in intent
	Nodes               []model.SentenceGroup
}

//GetFlowSetting gets the flow's all setting
func GetFlowSetting(nav int64, enterprise string) (*DetailNavFlow, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}

	isDelete := 0
	q := &model.NavQuery{ID: []int64{nav}, IsDelete: &isDelete, Enterprise: &enterprise}
	flow, err := navDao.GetFlows(dbLike.Conn(), q, nil)
	if err != nil {
		logger.Error.Printf("get flow failed. %s\n", err)
		return nil, err
	}
	if len(flow) == 0 {
		return nil, nil
	}

	var nodeOrder []string
	if flow[0].NodeOrder != "" {
		err = json.Unmarshal([]byte(flow[0].NodeOrder), &nodeOrder)
		if err != nil {
			return nil, fmt.Errorf("unmarshal node_order failed. %d. %s", flow[0].ID, err)
		}
	}
	senGrpsID, err := navDao.GetNodeID(dbLike.Conn(), []int64{nav})
	if err != nil {
		logger.Error.Printf("get node id failed")
		return nil, err
	}

	int64SenGrpIDs := make([]int64, 0, len(senGrpsID[nav])+1)
	for _, v := range senGrpsID[nav] {
		int64SenGrpIDs = append(int64SenGrpIDs, v)
	}
	if flow[0].IntentLinkID != 0 {
		int64SenGrpIDs = append(int64SenGrpIDs, flow[0].IntentLinkID)
	}

	resp := &DetailNavFlow{Nodes: []model.SentenceGroup{}, NavFlow: *flow[0]}

	if len(int64SenGrpIDs) > 0 {
		senGrps, err := sentenceGroupDao.GetNewBy(int64SenGrpIDs, nil, dbLike.Conn())
		if err != nil {
			logger.Error.Printf("get sentence group failed. %s\n", err)
			return nil, err
		}

		var intent model.SentenceGroup
		senGrpsMap := make(map[string]model.SentenceGroup)
		for _, senGrp := range senGrps {
			if senGrp.ID == flow[0].IntentLinkID {
				intent = senGrp
			} else {
				senGrpsMap[senGrp.UUID] = senGrp
			}
		}

		nodes := make([]model.SentenceGroup, 0, len(senGrpsMap))
		fixedOrder := make([]string, 0, len(nodeOrder))
		for _, uuid := range nodeOrder {
			if v, ok := senGrpsMap[uuid]; ok {
				nodes = append(nodes, v)
				delete(senGrpsMap, uuid)
				fixedOrder = append(fixedOrder, uuid)
			}
		}
		//append the rest node in case node_order is mess up
		for _, v := range senGrpsMap {
			nodes = append(nodes, v)
		}

		if len(fixedOrder) != len(nodeOrder) {
			updateNodeOrder(dbLike.Conn(), fixedOrder, flow[0].ID)
		}

		resp.Nodes = nodes
		resp.SentenceGroup = intent
	}
	return resp, nil
}

//GetFlows gets flows, but not include the node it includes
func GetFlows(q *model.NavQuery, page int, limit int) ([]*model.NavFlow, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}
	l := &model.NavLimit{Limit: limit, Page: page}
	flows, err := navDao.GetFlows(dbLike.Conn(), q, l)
	if err != nil {
		logger.Error.Printf("Get flows failed. %s\n", err)
	}
	return flows, err
}

//CountFlows counts the flows
func CountFlows(q *model.NavQuery) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	total, err := navDao.CountFlows(dbLike.Conn(), q)
	if err != nil {
		logger.Error.Printf("Count flows failed. %s\n", err)
	}
	return total, err
}

//CountNodes counts the node number in the given navs
//returnd value is the map with the key as the given nav and value is the count
func CountNodes(navs []int64) (map[int64]int64, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}
	if len(navs) == 0 {
		return make(map[int64]int64, 0), nil
	}
	resp, err := navDao.CountNodes(dbLike.Conn(), navs)
	if err != nil {
		logger.Error.Printf("Count nodes failed.%s\n", err)
	}

	return resp, err
}

//the Conversation type
const (
	AudioFile = iota
	Flow
)

func createFlowConversation(enterprise string, user string, body *apiFlowCreateBody) (string, error) {

	if dbLike == nil {
		return "", ErrNilCon
	}

	usingStat := MStatUsing
	models, err := modelDao.TrainedModelInfo(dbLike.Conn(),
		&model.TModelQuery{Status: &usingStat, Enterprise: &enterprise})

	if err != nil {
		logger.Error.Printf("get model failed. %s\n", err)
		return "", err
	}

	if len(models) == 0 {
		logger.Warn.Printf("enterprise %s has no trained model and tries to use navigation flow\n", enterprise)
		return "", ErrNoModels
	}

	tx, err := dbLike.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	reqCall := &NewCallReq{FileName: body.FileName, Enterprise: enterprise, Type: model.CallTypeRealTime, CallTime: body.CreateTime,
		UploadUser: user, LeftChannel: CallStaffRoleName, RightChannel: CallCustomerRoleName}
	call, err := NewCall(reqCall)
	if err != nil {
		logger.Error.Printf("create the call failed. %s\n", err)
		return "", err
	}

	settings, err := getCurSetting(enterprise)
	if err != nil {
		logger.Error.Printf("get flow setting %s failed. %s\n", enterprise, err)
		return "", err
	}

	settings.NavResult.FileName = body.FileName
	settings.NavResult.ID = call.UUID
	settings.CallID = call.ID
	settingsStr, err := json.Marshal(settings)
	if err != nil {
		logger.Error.Printf("Marshal failed. %s\n", err)
		return "", err
	}
	_, err = navOnTheFlyDao.InitConversationResult(tx, call.ID, int64(models[0].ID), string(settingsStr))
	if err != nil {
		logger.Error.Printf("insert empty flow result failed")
		return "", err
	}
	tx.Commit()

	return call.UUID, nil
}

//Error msg
var (
	ErrSpeaker        = errors.New("Wrong speaker")
	ErrEndTimeSmaller = errors.New("end time < start time")
)

//finishFlowQI finishs the flow, update information
func finishFlowQI(req *apiFlowFinish, uuid string) error {
	if dbLike == nil {
		return ErrNilCon
	}

	calls, err := Calls(dbLike.Conn(), model.CallQuery{UUID: []string{uuid}})
	if err != nil || len(calls) == 0 {
		logger.Error.Printf("Get conversation [%s] error. %s\n", uuid, err)
		return err
	}

	dur := req.FinishTime - calls[0].CallUnixTime
	if dur < 0 {
		return ErrEndTimeSmaller
	}

	tx, err := dbLike.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	calls[0].Status = model.CallStatusDone
	calls[0].DurationMillSecond = int(dur * 1000)

	err = callDao.SetCall(tx, calls[0])
	if err != nil {
		logger.Error.Printf("lupdate streaming conversation finished status failed. %s\n", err)
		return err
	}

	tx.Commit()

	return nil
}

type NavFlowSetting struct {
	Model     int64                        `json:"model"`
	NavResult NavResponse                  `json:"nav"`
	Criteria  map[uint64]*SenGroupCriteria `json:"critetia"`
	Levels    []map[uint64][]uint64        `json:"levels"`
	CallID    int64                        `json:"callID"`
	NodeLocal map[int64][]CreditLoc        `json:"postion"` //sentence group id to the location in the Flows
}

type NavResponse struct {
	FileName  string    `json:"file_name"`
	NavResult NavResult `json:"nav_result"`
	ID        string    `json:"id"`
}

type NavResult struct {
	Flows []StreamingFlow `json:"flows"`
}
type CreditLoc struct {
	IsIntent  bool `json:"is_intent"`
	FlowOrder int  `json:"flow_order"`
	NodeOrder int  `json:"node_order"`
}

type StreamingFlow struct {
	ID         string          `json:"nav_id"`
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	IntentName string          `json:"intent_name"`
	Role       string          `json:"role"`
	Valid      bool            `json:"valid"`
	Nodes      []StreamingNode `json:"nodes"`
}
type StreamingNode struct {
	ID       string `json:"sg_id"`
	Optional bool   `json:"optional"`
	Role     string `json:"role"`
	Name     string `json:"name"`
	Valid    bool   `json:"valid"`
}

type MatchedFlowNode struct {
	NavID    string `json:"nav_id"`
	Type     string `json:"type"`
	SenGrpID string `json:"sg_id"`
}

//this function would get the current flows and its nodes
//incluing the setting information, sentence group
func getCurSetting(enterprise string) (*NavFlowSetting, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}

	//get the trained model by the enterprise
	models, err := GetUsingModelByEnterprise(enterprise)
	numOfModels := len(models)
	if numOfModels == 0 {
		return nil, ErrNoModels
	} else if numOfModels > 1 {
		logger.Warn.Printf("More than 1 models is marked as using status with enterprise %s\n", enterprise)
		logger.Warn.Printf("Using the first one %d as prediction model", models[0].ID)
	}

	resp := &NavFlowSetting{Model: int64(models[0].ID), NodeLocal: make(map[int64][]CreditLoc)}

	//get the current navigation flows
	isDelete := 0
	q := &model.NavQuery{IsDelete: &isDelete, Enterprise: &enterprise}
	flows, err := navDao.GetFlows(dbLike.Conn(), q, nil)
	if err != nil {
		logger.Error.Printf("get flow failed. %s\n", err)
		return nil, err
	}
	if len(flows) == 0 {
		return nil, nil
	}

	allNewSenGrpsIDUint64 := make([]uint64, 0) //records the sentence group id, including the intent

	//copy the intent id if it is not fixed in the ui, intent is also a sentence group
	intentIDs := make([]int64, 0, len(flows))
	flowIDs := make([]int64, 0, len(flows))
	for _, v := range flows {
		if v.IgnoreIntent == 0 {
			intentIDs = append(intentIDs, v.IntentLinkID)
			allNewSenGrpsIDUint64 = append(allNewSenGrpsIDUint64, uint64(v.IntentLinkID))
		}
		flowIDs = append(flowIDs, v.ID)
	}

	//get all nodes' id, node is also the sentence group
	navIDToSenGrpsIDMap, err := navDao.GetNodeID(dbLike.Conn(), flowIDs)
	if err != nil {
		logger.Error.Printf("get node id failed")
		return nil, err
	}

	//append the intent id to node id, they are all sentence group
	allSenGrpsID := make([]int64, 0, len(navIDToSenGrpsIDMap)+len(flows))
	for _, v := range navIDToSenGrpsIDMap {
		allSenGrpsID = append(allSenGrpsID, v...)
	}
	allSenGrpsID = append(allSenGrpsID, intentIDs...)

	allSenGrpsIDUint64 := make([]uint64, 0, len(allSenGrpsID))
	for _, v := range allSenGrpsID {
		allSenGrpsIDUint64 = append(allSenGrpsIDUint64, uint64(v))
	}

	//extract the map[flowID][]string , flowID and its node's uuid
	sgFilter := &model.SentenceGroupFilter{ID: allSenGrpsIDUint64}
	_, mayOldSenGrps, err := GetSentenceGroupsBy(sgFilter)
	if err != nil {
		logger.Error.Printf("get sentence group failed.%s. %+v\n", err, *sgFilter)
		return nil, err
	}
	senGrpIDToUUIDMap := make(map[int64]string)
	for _, v := range mayOldSenGrps {
		senGrpIDToUUIDMap[v.ID] = v.UUID
	}

	//mkae a map, each flow has its node's uuid
	navIDToSenGrpsUUIDMap := make(map[int64]map[string]bool)
	for flowID, nodeIDs := range navIDToSenGrpsIDMap {
		for _, nodeID := range nodeIDs {
			if _, ok := navIDToSenGrpsUUIDMap[flowID]; !ok {
				navIDToSenGrpsUUIDMap[flowID] = make(map[string]bool)
			}
			uuid := senGrpIDToUUIDMap[nodeID]
			navIDToSenGrpsUUIDMap[flowID][uuid] = true
		}
	}

	//get the current sentence group, it may be different in id, but the same in uuid
	senGrps, err := sentenceGroupDao.GetNewBy(allSenGrpsID, nil, dbLike.Conn())
	if err != nil {
		logger.Error.Printf("get sentence group failed. %s\n", err)
		return nil, err
	}

	//lookup map by id and uuid
	senGrpMap := make(map[int64]model.SentenceGroup)
	senGrpUUIDMap := make(map[string]model.SentenceGroup)
	for _, v := range senGrps {
		senGrpMap[v.ID] = v
		senGrpUUIDMap[v.UUID] = v
		allNewSenGrpsIDUint64 = append(allNewSenGrpsIDUint64, uint64(v.ID))
	}

	//get the relation from sentence group to tag
	levels, _, err := GetLevelsRel(LevSenGroup, LevTag, allNewSenGrpsIDUint64)
	if err != nil {
		logger.Error.Printf("get level relations failed. %s\n", err)
		return nil, err
	}
	if len(levels) < 2 {
		err = fmt.Errorf("expected has 2 level relationn, but get %d", len(levels))
		logger.Error.Printf("%s", err)
		return nil, err
	}

	//create the sentence group criteria according to the current setting
	senGrpCriteria := make(map[uint64]*SenGroupCriteria)
	senGrpContainSen := levels[0]
	for i := 0; i < len(senGrps); i++ {
		id := uint64(senGrps[i].ID)
		var criterion SenGroupCriteria

		if senIDs, ok := senGrpContainSen[id]; ok {
			if senGrps[i].Optional != 1 {
				senGrpCriteria[id] = &criterion
				senGrpCriteria[id].ID = id
				senGrpCriteria[id].Role = senGrps[i].Role
				senGrpCriteria[id].Range = senGrps[i].Distance
				senGrpCriteria[id].Position = senGrps[i].Position
				senGrpCriteria[id].SentenceID = senIDs
			}
		} else {
			if senGrps[i].Optional != 1 {
				logger.Warn.Printf("No sentence group id %d in relation table, but exist in sentence group\n", id)
			}
			//return nil, ErrRequestNotEqualGet
		}
	}

	flowsSetting := make([]StreamingFlow, 0, len(flows))
	for flowIdx, flow := range flows {
		sf := StreamingFlow{Name: flow.Name, IntentName: flow.IntentName, ID: flow.UUID}
		if flow.IgnoreIntent == 0 {
			sf.Type = callInIntentCodeMap[1]
			loc := CreditLoc{FlowOrder: flowIdx, IsIntent: true}
			resp.NodeLocal[flow.IntentLinkID] = append(resp.NodeLocal[flow.IntentLinkID], loc)
		} else {
			sf.Type = callInIntentCodeMap[0]
		}
		if grp, ok := senGrpMap[flow.ID]; ok {
			sf.Role = roleCodeMap[grp.Role]
		}
		sf.Nodes = make([]StreamingNode, 0, len(navIDToSenGrpsUUIDMap[flow.ID]))

		var nodeOrder []string
		if flow.NodeOrder != "" {
			err = json.Unmarshal([]byte(flow.NodeOrder), &nodeOrder)
			if err != nil {
				err = fmt.Errorf("unmarshal node order failed. %d,%s. %s", flow.ID, flow.NodeOrder, err)
				logger.Error.Printf("%s\n", err)
				return nil, err
			}
		}

		nodeCounter := 0
		for _, uuid := range nodeOrder {
			if node, ok := senGrpUUIDMap[uuid]; ok {
				sn := StreamingNode{Name: node.Name, Optional: node.Optional == 1, Role: roleCodeMap[node.Role], ID: uuid}
				sf.Nodes = append(sf.Nodes, sn)
				loc := CreditLoc{FlowOrder: flowIdx, NodeOrder: nodeCounter}
				resp.NodeLocal[node.ID] = append(resp.NodeLocal[node.ID], loc)
				nodeCounter++
				delete(navIDToSenGrpsUUIDMap[flow.ID], uuid)

			}
		}
		for uuid := range navIDToSenGrpsUUIDMap[flow.ID] {
			if node, ok := senGrpUUIDMap[uuid]; ok {
				sn := StreamingNode{Name: node.Name, Optional: node.Optional == 1, Role: roleCodeMap[node.Role], ID: uuid}
				sf.Nodes = append(sf.Nodes, sn)
				loc := CreditLoc{FlowOrder: flowIdx, NodeOrder: nodeCounter}
				resp.NodeLocal[node.ID] = append(resp.NodeLocal[node.ID], loc)
				nodeCounter++
			}
		}

		flowsSetting = append(flowsSetting, sf)
	}

	resp.NavResult.NavResult.Flows = flowsSetting
	resp.Criteria = senGrpCriteria
	resp.Levels = levels
	return resp, nil
}

func updateStreamingPredict(callID int64, predict string) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}

	return navOnTheFlyDao.UpdateFlowResult(dbLike.Conn(), callID, predict)
}

func getStreamingPredict(callID int64) ([]*model.StreamingPredict, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}
	return navOnTheFlyDao.GetStreamingPredictResult(dbLike.Conn(), callID)
}

func streamingMatch(segs []*SegmentWithSpeaker, s *NavFlowSetting) ([]MatchedFlowNode, error) {

	numOfLines := len(segs)
	if numOfLines == 0 {
		return nil, nil
	}
	if s == nil {
		return nil, nil
	}
	if len(s.Levels) < 2 {
		return nil, ErrWrongLevel
	}
	if s.Criteria == nil {
		return nil, ErrNilSentenceGroup
	}

	//copy the text for each segments
	lines := make([]string, 0, numOfLines)
	for _, seg := range segs {
		lines = append(lines, seg.Text)
	}

	model := s.Model

	//calling cu model to check the matched tag
	timeout := time.Duration(30 * time.Second)
	tagMatchDat, err := TagMatch([]uint64{uint64(model)}, lines, timeout)
	if err != nil {
		logger.Error.Printf("doing tag matched failed. model:%d. %s\n",
			model, err)
		return nil, err
	}
	segMatchedTag := extractTagMatchedData(tagMatchDat)

	//doing sentence matched check
	senMatchDat, err := SentencesMatch(segMatchedTag, s.Levels[1])
	if err != nil {
		logger.Warn.Printf("doing sentence  match failed.model:%d. %s\n",
			model, err)
		return nil, err
	}

	//do the check, sentence group
	matchSgID, err := SentenceGroupMatch(senMatchDat, s.Criteria, segs)
	if err != nil {
		logger.Warn.Printf("doing sentence group match failed.%s\n", err)
		return nil, err
	}

	resp := make([]MatchedFlowNode, 0)
	//fill up the matched sentence group
	for matched := range matchSgID {
		for _, loc := range s.NodeLocal[int64(matched)] {
			navID := s.NavResult.NavResult.Flows[loc.FlowOrder].ID
			v := MatchedFlowNode{NavID: navID}
			if loc.IsIntent {
				s.NavResult.NavResult.Flows[loc.FlowOrder].Valid = true
				v.Type = "intent"
			} else {
				s.NavResult.NavResult.Flows[loc.FlowOrder].Nodes[loc.NodeOrder].Valid = true
				v.Type = "node"
				v.SenGrpID = s.NavResult.NavResult.Flows[loc.FlowOrder].Nodes[loc.NodeOrder].ID
			}
			resp = append(resp, v)
		}
	}
	return resp, nil
}
