package qi

import (
	"fmt"
	"time"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

// NewCondition create a default Condition with its groupID.
func NewCondition(cond model.Condition) (model.Condition, error) {
	return newCondition(nil, cond)
}

// NewCustomGroupConditions create a list of custom condition of the group by given customcolumns.
// customcolumns is a map contains filter value. As each key(UserKey) with slice of values(UserValue).
// ex:
//		customcolumns{
//			"Location": ["Taipei", "Taichung", "Shanghai"]
//		}
func NewCustomConditions(group model.Group, customcolumns map[string][]interface{}) ([]model.UserValue, error) {
	tx, err := dbLike.Begin()
	if err != nil {
		return nil, fmt.Errorf("new transaction from dbLike failed, %v", err)
	}
	uvs, err := newCustomConditions(tx, group, customcolumns)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return uvs, nil
}

func newCustomConditions(tx model.SQLTx, group model.Group, customcolumns map[string][]interface{}) ([]model.UserValue, error) {
	uvs := make([]model.UserValue, 0)
	timestamp := time.Now().Unix()
	for colName, values := range customcolumns {
		keys, err := userKeys(nil, model.UserKeyQuery{
			InputNames: []string{colName},
			Enterprise: group.EnterpriseID,
		})
		if err != nil {
			return nil, fmt.Errorf("query user key failed, %v", err)
		}
		if len(keys) == 0 {
			return nil, fmt.Errorf("user key '%s' does not exist", colName)
		}
		for value := range values {
			v := model.UserValue{
				UserKeyID:  keys[0].ID,
				LinkID:     group.ID,
				Type:       model.UserValueTypGroup,
				Value:      fmt.Sprintf("%d", value),
				CreateTime: timestamp,
				UpdateTime: timestamp,
			}
			v, err = newUserValue(tx, v)
			if err != nil {
				return nil, fmt.Errorf("new uservalue failed, %v", err)
			}
			uvs = append(uvs, v)
		}
	}
	return uvs, nil
}
