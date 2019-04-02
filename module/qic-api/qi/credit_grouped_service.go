package qi

import (
	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
)

//RetrieveGroupedCredit gets the grouped credit by call group id
func RetrieveGroupedCredit(callGroupUUID string) ([]*HistoryCredit, error) {
	if dbLike == nil {
		return nil, ErrNilCon
	}

	uuids := []string{callGroupUUID}
	cgQuery := model.CallGroupQuery{
		UUIDs: &uuids,
	}
	callGroupList, err := callGroupDao.GetCallGroups(dbLike.Conn(), &cgQuery)
	if err != nil {
		logger.Error.Printf("failed to get CallGroup list. %s\n", err)
		return nil, err
	}
	if len(callGroupList) == 0 {
		return []*HistoryCredit{}, nil
	}
	callGroupID := callGroupList[0].ID
	callIDs := callGroupList[0].Calls

	isGroup := len(callIDs) > 1
	credits := []*model.SimpleCredit{}
	if isGroup {
		query := model.CreditCallGroupQuery{
			CallGroupIDs: []uint64{uint64(callGroupID)},
		}
		creditCGs, err := creditCallGroupDao.GetCreditCallGroups(dbLike.Conn(), &query)
		if err != nil {
			logger.Error.Printf("get grouped credit failed\n")
			return nil, err
		}
		credits, err = convertToSimpleCredit(creditCGs)
		if err != nil {
			logger.Error.Printf("convert to SimpleCredit failed\n")
			return nil, err
		}
	} else {
		query := model.CreditQuery{
			Calls: []uint64{uint64(callIDs[0])},
		}
		credits, err = creditDao.GetCallCredit(dbLike.Conn(), &query)
		if err != nil {
			logger.Error.Printf("get Credit failed\n")
			return nil, err
		}
	}
	return buildHistroyCreditTree(callIDs, credits)
}

func convertToSimpleCredit(creditCGs []*model.CreditCallGroup) ([]*model.SimpleCredit, error) {
	credits := []*model.SimpleCredit{}
	for _, creditCG := range creditCGs {
		credit := &model.SimpleCredit{
			ID:         creditCG.ID,
			Type:       creditCG.Type,
			ParentID:   creditCG.ParentID,
			OrgID:      creditCG.OrgID,
			Valid:      creditCG.Valid,
			Revise:     creditCG.Revise,
			Score:      creditCG.Score,
			CreateTime: creditCG.CreateTime,
			UpdateTime: creditCG.UpdateTime,
		}
		credits = append(credits, credit)
	}
	return credits, nil
}
