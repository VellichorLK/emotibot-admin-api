package FAQ

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
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
			util.LogError.Printf("Error when parse row: %s", err.Error())
			return nil, err
		}
		obj := &Label{id, name, 0}
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

func getAllLabelActivityCount(appid string) (map[int]int, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT Label_Id, Count(*) FROM %s_activitylabel WHERE Status = 1 group by Label_Id", appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := map[int]int{}
	for rows.Next() {
		var id int
		var count int
		err := rows.Scan(&id, &count)
		if err != nil {
			util.LogError.Printf("Error when parse row: %s", err.Error())
			return nil, err
		}
		ret[id] = count
	}
	return ret, nil
}

func getLabelActivityCount(appid string, labelID int) (int, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return 0, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT Count(*) FROM %s_activitylabel WHERE Status = 1 and Label_Id = ? group by Label_Id", appid)
	row := mySQL.QueryRow(queryStr, labelID)

	var count int
	err := row.Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		util.LogError.Printf("Error when parse row: %s", err.Error())
		return 0, err
	}
	return count, nil
}

// =======================
// Start of activity part
// =======================

func addActivity(appid string, activity *Activity) (err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}
	// Use transition because activity may need to update two table at one time
	tx, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer func() {
		util.ClearTransition(tx)
	}()
	err = dbAddActivity(appid, activity, tx)
	if err != nil {
		return
	}
	if activity.LinkLabel != nil {
		err = dbLinkActivityLabel(appid, activity.ID, *activity.LinkLabel, tx)
	}
	if err != nil {
		return
	}
	err = tx.Commit()
	return
}
func updateActivity(appid string, newActivity *Activity) (err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}
	// Use transition because activity may need to update two table at one time
	tx, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer func() {
		util.ClearTransition(tx)
	}()
	err = dbUpdateActivity(appid, newActivity, tx)
	if err != nil {
		return
	}
	err = dbUnlinkActivityLabel(appid, newActivity.ID, tx)
	if err != nil {
		return
	}
	if newActivity.LinkLabel != nil {
		err = dbLinkActivityLabel(appid, newActivity.ID, *newActivity.LinkLabel, tx)
	}
	if err != nil {
		return
	}
	err = tx.Commit()
	return
}
func deleteActivity(appid string, id int) (err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}
	// Use transition because activity may need to update two table at one time
	tx, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer func() {
		util.ClearTransition(tx)
	}()
	err = dbDeleteActivity(appid, id, tx)
	if err != nil {
		return
	}
	err = dbUnlinkActivityLabel(appid, id, tx)
	if err != nil {
		return
	}
	err = tx.Commit()
	return
}

func dbDeleteActivity(appid string, id int, tx *sql.Tx) error {
	queryStr := fmt.Sprintf("DELETE FROM %s_activity WHERE Activity_Id = ?", appid)
	ret, err := tx.Exec(queryStr, id)
	if err != nil {
		return err
	}
	_, err = ret.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}
func dbAddActivity(appid string, activity *Activity, tx *sql.Tx) error {
	queryStr := fmt.Sprintf(`
		INSERT INTO %s_activity (Name, Content, Status, Begin_Time, End_Time)
		VALUES (?, ?, ?, ?, ?)`, appid)
	ret, err := tx.Exec(queryStr,
		activity.Name, activity.Content, activity.Status,
		activity.StartTime, activity.EndTime)
	if err != nil {
		return err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return err
	}
	activity.ID = int(id)
	return nil
}
func dbUpdateActivity(appid string, newActivity *Activity, tx *sql.Tx) error {
	queryStr := fmt.Sprintf(`
		UPDATE %s_activity
		SET Name = ?, Content = ?, Status = ?, Begin_Time = ?, End_Time = ?
		WHERE Activity_Id = ?`, appid)
	ret, err := tx.Exec(queryStr,
		newActivity.Name, newActivity.Content, newActivity.Status,
		newActivity.StartTime, newActivity.EndTime, newActivity.ID)
	if err != nil {
		return err
	}
	_, err = ret.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}
func dbUnlinkActivityLabel(appid string, aid int, tx *sql.Tx) error {
	queryStr := fmt.Sprintf(`DELETE FROM %s_activitylabel WHERE Activity_Id = ?`, appid)
	_, err := tx.Exec(queryStr, aid)
	return err
}
func dbLinkActivityLabel(appid string, aid int, lid int, tx *sql.Tx) error {
	queryStr := fmt.Sprintf(`INSERT INTO %s_activitylabel (Activity_Id, Label_Id) VALUES (?, ?)`, appid)
	_, err := tx.Exec(queryStr, aid, lid)
	return err
}

func setActivityStatus(appid string, id int, status bool) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	input := 0
	if status {
		input = 1
	}
	queryStr := fmt.Sprintf(`
		UPDATE %s_activity SET Status = ? WHERE Activity_Id = ?`, appid)
	ret, err := mySQL.Exec(queryStr, input, id)
	if err != nil {
		return err
	}
	_, err = ret.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

func unlinkActivityLabel(appid string, aid int) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}
	queryStr := fmt.Sprintf(`DELETE FROM %s_activitylabel WHERE Activity_Id = ?`, appid)
	_, err := mySQL.Exec(queryStr, aid)
	return err
}
func linkActivityLabel(appid string, aid int, lid int) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}
	queryStr := fmt.Sprintf(`INSERT INTO %s_activitylabel (Activity_Id, Label_Id) VALUES (?, ?)`, appid)
	_, err := mySQL.Exec(queryStr, aid, lid)
	return err
}

func getActivities(appid string) ([]*Activity, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf(`
		SELECT 
			Activity_Id, Name, Content, Begin_Time, End_Time, Status
		FROM %s_activity`, appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []*Activity{}
	for rows.Next() {
		obj := Activity{}
		var status int
		err := rows.Scan(&obj.ID, &obj.Name, &obj.Content, &obj.StartTime, &obj.EndTime, &status)
		obj.Status = status != 0
		if err != nil {
			util.LogError.Printf("Error when parse activity row: %s", err.Error())
			return nil, err
		}

		ret = append(ret, &obj)
	}
	return ret, nil
}

func getActivityByID(appid string, id int) (*Activity, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf(`
		SELECT 
			Activity_Id, Name, Content, Begin_Time, End_Time, Status
		FROM %s_activity
		WHERE Activity_Id = ?`, appid)
	row := mySQL.QueryRow(queryStr, id)
	ret := Activity{}
	var status int
	err := row.Scan(&ret.ID, &ret.Name, &ret.Content, &ret.StartTime, &ret.EndTime, &status)
	if err != nil {
		util.LogError.Printf("Error when parse activity row: %s", err.Error())
		return nil, err
	}
	ret.Status = status != 0

	return &ret, nil
}

func getActivityLabel(appid string) (map[int]int, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT Activity_Id, Label_Id FROM %s_activitylabel WHERE Status = 1", appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := map[int]int{}
	for rows.Next() {
		var aid int
		var lid int
		err := rows.Scan(&aid, &lid)
		if err != nil {
			util.LogError.Printf("Error when parse row: %s", err.Error())
			return nil, err
		}
		ret[aid] = lid
	}
	return ret, nil
}

func setActivityLabel(appid string, aid int, lid int) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}
	queryStr := fmt.Sprintf(`UPDATE %s_activitylabel SET Label_Id = ? WHERE Activity_Id = ?`, appid)
	_, err := mySQL.Exec(queryStr, lid, aid)
	return err
}

func getActivityOfLabel(appid string, lid int) (int, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return 0, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT Activity_Id FROM %s_activitylabel WHERE Status = 1 AND Label_Id = ?", appid)
	row := mySQL.QueryRow(queryStr, lid)

	var aid int
	err := row.Scan(&aid)
	if err != nil {
		util.LogError.Printf("Error when parse row: %s", err.Error())
		return 0, err
	}
	return aid, nil
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

	return rules, nil
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
	result, err := mySQL.Exec(queryStr, queryParams...)
	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(id), err
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
	_, err = mySQL.Exec(queryStr, queryParams...)
	if err != nil {
		return err
	}

	return err
}
