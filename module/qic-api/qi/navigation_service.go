package qi

import (
	"time"

	"emotibot.com/emotigo/pkg/logger"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
)

var (
	navDao model.NavigationDao = &model.NavigationSQLDao{}
)

//NewFlow creates the new flow and sets the empty node and intent
func NewFlow(name string, enterprise string) (int64, error) {
	if dbLike == nil {
		return 0, ErrNilCon
	}
	uuid, err := general.UUID()
	if err != nil {
		logger.Error.Printf("generate uuid failed. %s\n", err)
		return 0, err
	}
	now := time.Now().Unix()
	q := &model.NavFlow{Name: name, Enterprise: enterprise,
		CreateTime: now, UpdateTime: now, IgnoreIntent: 1,
		UUID: uuid}
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
	err = tx.Commit()
	if err != nil {
		logger.Error.Printf("commit data failed. %s\n", err)
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
		intentIdx := -1
		for idx, senGrp := range senGrps {
			if senGrp.ID == flow[0].IntentLinkID {
				intentIdx = idx
				intent = senGrp
				break
			}
		}
		if intentIdx >= 0 {
			senGrps = append(senGrps[:intentIdx], senGrps[intentIdx+1:]...)
		}
		resp.Nodes = senGrps
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

func UpdateFlowIntent(nav int64, senGrp *model.SentenceGroup) error {
	return nil
}
