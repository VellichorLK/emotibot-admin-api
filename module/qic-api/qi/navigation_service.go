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
	senGrpsID, err := navDao.GetNodeID(dbLike.Conn(), nav)
	if err != nil {
		logger.Error.Printf("get node id failed")
		return nil, err
	}

	int64SenGrpIDs := make([]int64, 0, len(senGrpsID)+1)
	for _, v := range senGrpsID {
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

	empty := &model.QIFlowResult{FileName: body.FileName}
	emptStr, err := json.Marshal(empty)
	if err != nil {
		logger.Error.Printf("Marshal failed. %s\n", err)
		return "", err
	}
	_, err = navOnTheFlyDao.InitConerationResult(tx, call.ID, int64(models[0].ID), string(emptStr))
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
func finishFlowQI(req *apiFlowFinish, id int64, result *model.QIFlowResult) error {
	if dbLike == nil {
		return ErrNilCon
	}

	calls, err := Calls(dbLike.Conn(), model.CallQuery{ID: []int64{id}})
	if err != nil || len(calls) == 0 {
		logger.Error.Printf("Get conversation [%d] error. %s\n", id, err)
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

	resultStr, err := json.Marshal(result)
	if err != nil {
		logger.Error.Printf("Marshal failed. %s\n", err)
		return err
	}
	_, err = navOnTheFlyDao.UpdateFlowResult(tx, calls[0].ID, string(resultStr))
	if err != nil {
		logger.Error.Printf("lupdate flow result failed. %s\n", err)
		return err
	}

	calls[0].Status = 1
	calls[0].DurationMillSecond = int(dur * 1000)

	err = callDao.SetCall(tx, calls[0])
	if err != nil {
		logger.Error.Printf("lupdate streaming conversation finished status failed. %s\n", err)
		return err
	}

	tx.Commit()

	return nil
}

func streamingMatch([]model.AsrContent) {
	//	var s []string
}
