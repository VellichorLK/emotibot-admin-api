package FAQ

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

func getQuestionLabels(appid string) ([]*Label, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT Label_Id, Label_Name FROM %s_label WHERE Status = 1", appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []*Label{}
	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			logger.Error.Printf("Error when parse row: %s", err.Error())
			return nil, err
		}
		obj := &Label{ID: id, Name: name}
		ret = append(ret, obj)
	}
	return ret, nil
}

func getQuestionLabelByName(appid string, name string) (*Label, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT Label_Id, Label_Name FROM %s_label WHERE Status = 1 and Label_Name = ?", appid)
	row := mySQL.QueryRow(queryStr, name)

	ret := Label{}
	err := row.Scan(&ret.ID, &ret.Name)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func getQuestionLabelByID(appid string, id int) (*Label, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT Label_Id, Label_Name FROM %s_label WHERE Status = 1 and Label_Id = ?", appid)
	row := mySQL.QueryRow(queryStr, id)

	ret := Label{}
	err := row.Scan(&ret.ID, &ret.Name)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func addQuestionLabel(appid string, label *Label) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("INSERT into %s_label (Label_Name) VALUES (?)", appid)
	ret, err := mySQL.Exec(queryStr, label.Name)
	if err != nil {
		return err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return err
	}
	label.ID = int(id)
	return nil
}

func updateQuestionLabel(appid string, label *Label) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("UPDATE %s_label SET Label_Name = ? WHERE Label_Id = ?", appid)
	ret, err := mySQL.Exec(queryStr, label.Name, label.ID)
	if err != nil {
		return err
	}
	_, err = ret.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

func deleteQuestionLabel(appid string, id int) error {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("DELETE FROM %s_label WHERE Label_Id = ?", appid)
	ret, err := mySQL.Exec(queryStr, id)
	if err != nil {
		return err
	}
	_, err = ret.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

// =======================
// Start of Rule part
// =======================

func getRules(appid string) ([]*Rule, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return nil, err
	}

	queryStr := fmt.Sprintf(`
		SELECT
			rule_id, name, target, rule, answer,
			response_type, status, begin_time, end_time
		FROM %s_rule`, appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := []*Rule{}
	for rows.Next() {
		temp := &Rule{}
		err = scanRowToRule(rows, temp)
		if err != nil {
			fmt.Printf("Err: %s\n", err.Error())
			return nil, err
		}
		if temp != nil {
			rules = append(rules, temp)
		}
	}

	queryStr = fmt.Sprintf("SELECT rule_id, label_id FROM %s_rule_label", appid)
	idRows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer idRows.Close()

	idMap := map[int][]int{}
	for idRows.Next() {
		rid, lid := 0, 0
		err = idRows.Scan(&rid, &lid)
		if err != nil {
			return nil, err
		}
		if _, ok := idMap[rid]; !ok {
			idMap[rid] = []int{}
		}
		idMap[rid] = append(idMap[rid], lid)
	}

	for _, rule := range rules {
		if ids, ok := idMap[rule.ID]; ok {
			rule.LinkLabel = ids
		} else {
			rule.LinkLabel = []int{}
		}
	}

	return rules, err
}

func scanRowToRule(rows *sql.Rows, temp *Rule) error {
	ruleStr := ""
	err := rows.Scan(&temp.ID, &temp.Name, &temp.Target, &ruleStr, &temp.Answer,
		&temp.Type, &temp.Status, &temp.Begin, &temp.End)
	if err != nil {
		return err
	}

	ruleStr = strings.Replace(ruleStr, "\n", "", -1)
	ruleContents := []*RuleContent{}
	err = json.Unmarshal([]byte(ruleStr), &ruleContents)
	if err != nil {
		fmt.Printf("Error json: \n%s\n\n", ruleStr)
		return fmt.Errorf("Invalid rule content: %s", err.Error())
	}

	temp.Rule = ruleContents
	temp.Answer = strings.Replace(temp.Answer, "\n", "", -1)
	temp.Answer = strings.Replace(temp.Answer, "\t", "", -1)
	temp.Answer = strings.Replace(temp.Answer, "\r", "", -1)
	return err
}

func getRule(appid string, id int) (*Rule, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return nil, err
	}

	queryStr := fmt.Sprintf(`
		SELECT
			rule_id, name, target, rule, answer,
			response_type, status, begin_time, end_time
		FROM %s_rule
		WHERE rule_id = ?`, appid)
	rows, err := mySQL.Query(queryStr, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		rule := &Rule{}
		err = scanRowToRule(rows, rule)
		if err != nil {
			fmt.Printf("Err: %s\n", err.Error())
			return nil, err
		}
		return rule, nil
	}

	return nil, nil
}

func deleteRule(appid string, id int) error {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return err
	}

	queryStr := fmt.Sprintf(`
		DELETE
		FROM %s_rule
		WHERE rule_id = ?`, appid)
	_, err = mySQL.Exec(queryStr, id)
	return err
}

func addRule(appid string, rule *Rule) (int, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return -1, err
	}

	tx, err := mySQL.Begin()
	if err != nil {
		return -1, err
	}
	defer util.ClearTransition(tx)

	queryStr := fmt.Sprintf(`
		INSERT INTO %s_rule
		(name, target, rule, answer, response_type, status, begin_time, end_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, appid)
	ruleStr, _ := json.Marshal(rule.Rule)
	statusInt := 0
	if rule.Status {
		statusInt = 1
	}
	queryParams := []interface{}{
		rule.Name,
		rule.Target,
		ruleStr,
		rule.Answer,
		rule.Type,
		statusInt,
		rule.Begin,
		rule.End,
	}
	result, err := tx.Exec(queryStr, queryParams...)
	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	queryStr = fmt.Sprintf(`
		INSERT INTO %s_rule_label
		(rule_id, label_id) VALUES (?, ?)`, appid)
	for _, labelID := range rule.LinkLabel {
		_, err := tx.Exec(queryStr, id, labelID)
		if err != nil {
			return -1, err
		}
	}

	return int(id), tx.Commit()
}

func updateRule(appid string, id int, rule *Rule) error {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return err
	}

	tx, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(tx)

	queryStr := fmt.Sprintf(`
		UPDATE %s_rule SET
		name = ?, target = ?, rule = ?, answer = ?, 
		response_type = ?, status = ?, begin_time = ?, end_time = ?
		WHERE rule_id = ?`, appid)
	ruleStr, _ := json.Marshal(rule.Rule)
	statusInt := 0
	if rule.Status {
		statusInt = 1
	}
	queryParams := []interface{}{
		rule.Name,
		rule.Target,
		ruleStr,
		rule.Answer,
		rule.Type,
		statusInt,
		rule.Begin,
		rule.End,
		id,
	}
	_, err = tx.Exec(queryStr, queryParams...)
	if err != nil {
		return err
	}

	queryStr = fmt.Sprintf(`
		DELETE FROM %s_rule_label
		WHERE rule_id = ?`, appid)
	_, err = tx.Exec(queryStr, id)
	if err != nil {
		return err
	}

	queryStr = fmt.Sprintf(`
		INSERT INTO %s_rule_label
		(rule_id, label_id) VALUES (?, ?)`, appid)
	for _, labelID := range rule.LinkLabel {
		_, err := tx.Exec(queryStr, id, labelID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func getRulesOfLabel(appid string, labelID int) ([]*Rule, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return nil, err
	}

	queryStr := fmt.Sprintf(`
		SELECT
			r.rule_id, r.name, r.target, r.rule, r.answer,
			r.response_type, r.status, r.begin_time, r.end_time
		FROM %s_rule as r, %s_rule_label as rl
		WHERE r.rule_id = rl.rule_id AND rl.label_id = ?`, appid, appid)
	rows, err := mySQL.Query(queryStr, labelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := []*Rule{}
	for rows.Next() {
		temp := &Rule{}
		err = scanRowToRule(rows, temp)
		if err != nil {
			fmt.Printf("Err: %s\n", err.Error())
			return nil, err
		}
		if temp != nil {
			rules = append(rules, temp)
		}
	}

	return rules, err
}

func getLabelsOfRule(appid string, ruleID int) ([]*Label, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf(`
		SELECT l.Label_Id, l.Label_Name
		FROM %s_label as l, %s_rule_label as rl
		WHERE Status = 1 AND rl.rule_id = ? AND rl.label_id = l.Label_Id`, appid, appid)
	rows, err := mySQL.Query(queryStr, ruleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []*Label{}
	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			logger.Error.Printf("Error when parse row: %s", err.Error())
			return nil, err
		}
		obj := &Label{ID: id, Name: name}
		ret = append(ret, obj)
	}
	return ret, nil
}

func getRuleCountOfLabels(appid string) (map[int]int, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	ret := map[int]int{}
	queryStr := fmt.Sprintf(
		`SELECT label_id, count(*)
		FROM %s_rule_label
		GROUP BY label_id`, appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return ret, err
	}

	for rows.Next() {
		id, count := 0, 0
		err = rows.Scan(&id, &count)
		if err != nil {
			return map[int]int{}, err
		}
		ret[id] = count
	}
	return ret, nil
}

func getLabelRuleCount(appid string, id int) (int, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return 0, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf(
		`SELECT count(*)
		FROM %s_rule_label
		WHERE label_id = ?
		GROUP BY label_id`, appid)
	row := mySQL.QueryRow(queryStr, id)
	count := 0
	err = row.Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return count, nil
}

func getLabelRuleCountMap(appid string) (map[int]int, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		return map[int]int{}, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf(
		`SELECT label_id, count(*)
		FROM %s_rule_label
		GROUP BY label_id`, appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil && err != sql.ErrNoRows {
		return map[int]int{}, err
	}
	defer rows.Close()

	ret := map[int]int{}
	for rows.Next() {
		id := 0
		count := 0
		err = rows.Scan(&id, &count)
		if _, ok := ret[id]; !ok {
			ret[id] = 0
		}
		ret[id]++
	}
	return ret, nil
}
