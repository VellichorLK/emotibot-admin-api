package FAQ

import (
	"database/sql"
	"errors"

	"emotibot.com/emotigo/module/admin-api/ApiError"
)

func GetQuestionLabels(appid string) ([]*Label, error) {
	labels, err := getQuestionLabels(appid)
	if err != nil {
		return nil, err
	}

	return labels, nil
}

func AddNewLabel(appid string, newLabel *Label) (int, error) {
	_, err := getQuestionLabelByName(appid, newLabel.Name)
	if err == nil {
		return ApiError.REQUEST_ERROR, errors.New("label existed")
	}
	err = addQuestionLabel(appid, newLabel)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func UpdateLabel(appid string, newLabel *Label) (int, error) {
	_, err := getQuestionLabelByID(appid, newLabel.ID)
	if err != nil {
		return ApiError.REQUEST_ERROR, errors.New("label not existed")
	}
	_, err = getQuestionLabelByName(appid, newLabel.Name)
	if err == nil {
		return ApiError.REQUEST_ERROR, errors.New("new name existed")
	}
	err = updateQuestionLabel(appid, newLabel)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func DeleteLabel(appid string, id int) (int, error) {
	_, err := getQuestionLabelByID(appid, id)
	if err != nil {
		return ApiError.REQUEST_ERROR, errors.New("label not existed")
	}
	count, err := getLabelActivityCount(appid, id)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	if count > 0 {
		return ApiError.REQUEST_ERROR, errors.New("Label link to activities")
	}
	err = deleteQuestionLabel(appid, id)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func GetRules(appid string) ([]*Rule, error) {
	return getRules(appid)
}

func GetRulesOfLabel(appid string, labelID int) ([]*Rule, error) {
	return getRulesOfLabel(appid, labelID)
}

func GetRule(appid string, id int) (*Rule, error) {
	return getRule(appid, id)
}

func DeleteRule(appid string, id int) error {
	err := deleteRule(appid, id)
	if err == sql.ErrNoRows {
		return nil
	}

	return err
}

func AddRule(appid string, rule *Rule) (int, error) {
	return addRule(appid, rule)
}

func UpdateRule(appid string, id int, rule *Rule) error {
	return updateRule(appid, id, rule)
}

func GetLabelsOfRule(appid string, ruleID int) ([]*Label, error) {
	labels, err := getLabelsOfRule(appid, ruleID)
	if err != nil {
		return nil, err
	}
	countMap, err := GetRuleCountOfLabels(appid)
	if err != nil {
		return nil, err
	}
	for _, l := range labels {
		if count, ok := countMap[l.ID]; ok {
			l.RuleCount = count
		}
	}
	return labels, nil
}

func GetRuleCountOfLabels(appid string) (map[int]int, error) {
	return getRuleCountOfLabels(appid)
}
