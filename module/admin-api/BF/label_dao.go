package BF

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

var errDuplicate = errors.New("duplicate item")
var errDBNotInit = errors.New("DB not init")

// =======================
// Start of Cmd part
// =======================

type scanner interface {
	Scan(...interface{}) error
}

func getCmds(appid string) (*CmdClass, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return nil, err
	}

	queryStr := "SELECT id, name, parent FROM cmd_class WHERE appid = ? ORDER BY id DESC"
	classRows, err := mySQL.Query(queryStr, appid)
	if err != nil {
		return nil, err
	}
	defer classRows.Close()

	root := CmdClass{
		ID:       -1,
		Name:     "",
		Cmds:     []*Cmd{},
		Children: []*CmdClass{},
	}
	classMap := map[int]*CmdClass{}
	classes := []*CmdClass{}
	parentMap := map[int]*int{}

	classMap[-1] = &root
	// parse all class first
	for classRows.Next() {
		temp := CmdClass{
			Cmds:     []*Cmd{},
			Children: []*CmdClass{},
		}
		var parent *int
		classRows.Scan(&temp.ID, &temp.Name, &parent)
		classMap[temp.ID] = &temp
		classes = append(classes, &temp)
		parentMap[temp.ID] = parent
	}
	// append to each class children by map
	for idx := range classes {
		class := classes[idx]
		parentPtr := parentMap[class.ID]
		// default parent is root
		parentID := -1
		if parentPtr != nil {
			parentID = *parentPtr
		}
		// if parent not existed, change it parent to root
		if parent, ok := classMap[parentID]; ok {
			parent.Children = append(parent.Children, class)
		} else {
			root.Children = append(root.Children, class)
		}
	}

	queryStr = `
		SELECT
			cid, cmd_id, name, target, rule, answer,
			response_type, status, begin_time, end_time
		FROM cmd WHERE appid = ? ORDER BY cmd_id DESC`
	rows, err := mySQL.Query(queryStr, appid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cmds := []*Cmd{}
	for rows.Next() {
		cidPtr, temp, err := scanRowToCmd(rows)
		if err != nil {
			fmt.Printf("Err: %s\n", err.Error())
			return nil, err
		}

		if temp == nil {
			continue
		}

		cid := -1
		if cidPtr != nil {
			cid = *cidPtr
		}
		if class, ok := classMap[cid]; ok {
			class.Cmds = append(class.Cmds, temp)
		} else {
			root.Cmds = append(root.Cmds, temp)
		}
		cmds = append(cmds, temp)
	}

	queryStr = `
		SELECT rl.cmd_id, rl.robot_tag_id
		FROM cmd_robot_tag AS rl, cmd AS r
		WHERE rl.cmd_id = r.cmd_id AND r.appid = ?`
	idRows, err := mySQL.Query(queryStr, appid)
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

	for _, cmd := range cmds {
		if ids, ok := idMap[cmd.ID]; ok {
			cmd.LinkLabel = ids
		} else {
			cmd.LinkLabel = []int{}
		}
	}

	return &root, err
}

func scanRowToCmd(rows scanner) (parentPtr *int, ret *Cmd, err error) {
	ruleStr := ""
	temp := &Cmd{}
	err = rows.Scan(&parentPtr, &temp.ID, &temp.Name, &temp.Target, &ruleStr, &temp.Answer,
		&temp.Type, &temp.Status, &temp.Begin, &temp.End)
	if err != nil {
		return
	}

	ruleStr = strings.Replace(ruleStr, "\n", "", -1)
	ruleContents := []*CmdContent{}
	err = json.Unmarshal([]byte(ruleStr), &ruleContents)
	if err != nil {
		fmt.Printf("Error json: \n%s\n\n", ruleStr)
		err = fmt.Errorf("Invalid rule content: %s", err.Error())
		return
	}

	temp.Rule = ruleContents
	temp.Answer = strings.Replace(temp.Answer, "\n", "", -1)
	temp.Answer = strings.Replace(temp.Answer, "\t", "", -1)
	temp.Answer = strings.Replace(temp.Answer, "\r", "", -1)
	ret = temp

	if parentPtr != nil && *parentPtr == -1 {
		parentPtr = nil
	}
	return
}

func getCmd(appid string, id int) (*Cmd, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return nil, err
	}

	queryStr := `
		SELECT
			cid, cmd_id, name, target, rule, answer,
			response_type, status, begin_time, end_time
		FROM cmd
		WHERE cmd_id = ? AND appid = ?`
	row := mySQL.QueryRow(queryStr, id, appid)

	_, cmd, err := scanRowToCmd(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	queryStr = `
		SELECT rl.robot_tag_id
		FROM cmd_robot_tag AS rl, cmd AS r
		WHERE rl.cmd_id = r.cmd_id AND r.appid = ? AND r.cmd_id = ?`
	idRows, err := mySQL.Query(queryStr, appid, id)
	if err != nil {
		return nil, err
	}
	defer idRows.Close()
	for idRows.Next() {
		lid := 0
		err = idRows.Scan(&lid)
		if err != nil {
			return nil, err
		}
		cmd.LinkLabel = append(cmd.LinkLabel, lid)
	}

	return cmd, nil
}

func deleteCmd(appid string, id int) error {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return err
	}

	queryStr := `
		DELETE
		FROM cmd
		WHERE cmd_id = ? AND appid = ?`
	_, err = mySQL.Exec(queryStr, id, appid)
	return err
}

func addCmd(appid string, cmd *Cmd, cid int) (int, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return -1, err
	}

	tx, err := mySQL.Begin()
	if err != nil {
		return -1, err
	}
	defer util.ClearTransition(tx)

	queryStr := `
		INSERT INTO cmd
		(cid, name, target, rule, answer, response_type, status, begin_time, end_time, appid)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	ruleStr, _ := json.Marshal(cmd.Rule)
	statusInt := 0
	if cmd.Status {
		statusInt = 1
	}
	cidPtr := &cid
	if cid == -1 {
		cidPtr = nil
	}
	queryParams := []interface{}{
		cidPtr,
		cmd.Name,
		cmd.Target,
		ruleStr,
		cmd.Answer,
		cmd.Type,
		statusInt,
		cmd.Begin,
		cmd.End,
		appid,
	}
	result, err := tx.Exec(queryStr, queryParams...)
	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	queryStr = `
		INSERT INTO cmd_robot_tag
		(cmd_id, robot_tag_id) VALUES (?, ?)`
	for _, labelID := range cmd.LinkLabel {
		_, err := tx.Exec(queryStr, id, labelID)
		if err != nil {
			return -1, err
		}
	}

	return int(id), tx.Commit()
}

func updateCmd(appid string, id int, cmd *Cmd) error {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return err
	}

	tx, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(tx)
	statusVal := 0
	if cmd.Status {
		statusVal = 1
	}

	queryStr := `
		UPDATE cmd SET
		name = ?, target = ?, rule = ?, answer = ?,
		response_type = ?, status = ?, begin_time = ?, end_time = ?,
		status = ?
		WHERE cmd_id = ? AND appid = ?`
	ruleStr, _ := json.Marshal(cmd.Rule)
	statusInt := 0
	if cmd.Status {
		statusInt = 1
	}
	queryParams := []interface{}{
		cmd.Name,
		cmd.Target,
		ruleStr,
		cmd.Answer,
		cmd.Type,
		statusInt,
		cmd.Begin,
		cmd.End,
		statusVal,
		id,
		appid,
	}
	_, err = tx.Exec(queryStr, queryParams...)
	if err != nil {
		return err
	}

	queryStr = `
		DELETE FROM cmd_robot_tag
		WHERE cmd_id = ?`
	_, err = tx.Exec(queryStr, id)
	if err != nil {
		return err
	}

	queryStr = `
		INSERT INTO cmd_robot_tag
		(cmd_id, robot_tag_id) VALUES (?, ?)`
	for _, labelID := range cmd.LinkLabel {
		_, err := tx.Exec(queryStr, id, labelID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func getCmdsOfLabel(appid string, labelID int) ([]*Cmd, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return nil, err
	}

	queryStr := `
		SELECT
			r.cid, r.cmd_id, r.name, r.target, r.cmd, r.answer,
			r.response_type, r.status, r.begin_time, r.end_time
		FROM cmd as r, cmd_robot_tag as rl
		WHERE r.cmd_id = rl.cmd_id AND rl.robot_tag_id = ? AND r.appid = ?`
	rows, err := mySQL.Query(queryStr, labelID, appid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cmds := []*Cmd{}
	for rows.Next() {
		_, temp, err := scanRowToCmd(rows)
		if err != nil {
			fmt.Printf("Err: %s\n", err.Error())
			return nil, err
		}
		if temp != nil {
			cmds = append(cmds, temp)
		}
	}

	return cmds, err
}

func getLabelsOfCmd(appid string, cmdID int) ([]*Label, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errDBNotInit
	}

	queryStr := `
		SELECT l.id, l.name
		FROM tbl_robot_tag as l, cmd_robot_tag as rl
		WHERE Status = 1 AND rl.cmd_id = ? AND rl.robot_tag_id = l.id AND
			rl.appid = l.appid AND rl.appid = ?`
	rows, err := mySQL.Query(queryStr, cmdID, appid)
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
		obj := &Label{ID: fmt.Sprintf("%d", id), Name: name}
		ret = append(ret, obj)
	}
	return ret, nil
}

func getCmdCountOfLabels(appid string) (map[int]int, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errDBNotInit
	}

	ret := map[int]int{}
	queryStr := `
		SELECT robot_tag_id, count(*)
		FROM cmd_robot_tag
		GROUP BY robot_tag_id`
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return ret, err
	}
	defer rows.Close()

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

func getLabelCmdCount(appid string, id int) (int, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return 0, errDBNotInit
	}

	queryStr := `
		SELECT count(*)
		FROM cmd_robot_tag
		WHERE robot_tag_id = ?
		GROUP BY robot_tag_id`
	row := mySQL.QueryRow(queryStr, id)
	count := 0
	err = row.Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return count, nil
}

func getLabelCmdCountMap(appid string) (map[int]int, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		return map[int]int{}, errDBNotInit
	}

	queryStr := `
		SELECT robot_tag_id, count(*)
		FROM cmd_robot_tag
		GROUP BY robot_tag_id`
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

func getCmdClass(appid string, classID int) (ret *CmdClass, err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return
	}

	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := `SELECT name FROM cmd_class WHERE appid = ? AND id = ?`
	row := t.QueryRow(queryStr, appid, classID)
	cmdClass := CmdClass{}
	cmdClass.Cmds = []*Cmd{}
	cmdClass.Children = []*CmdClass{}
	err = row.Scan(&cmdClass.Name)
	if err != nil {
		return
	}

	queryStr = `
		SELECT rl.cmd_id, rl.robot_tag_id
		FROM cmd_robot_tag AS rl, cmd AS r
		WHERE rl.cmd_id = r.cmd_id AND r.appid = ?`
	idRows, err := mySQL.Query(queryStr, appid)
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

	queryStr = `
		SELECT
			cid, cmd_id, name, target, rule, answer,
			response_type, status, begin_time, end_time
		FROM cmd WHERE appid = ? AND cid = ? ORDER BY cmd_id DESC`
	rows, err := t.Query(queryStr, appid, classID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			return
		}
	}
	defer rows.Close()

	for rows.Next() {
		var cmd *Cmd
		_, cmd, err = scanRowToCmd(rows)
		if err != nil {
			return
		}
		if ids, ok := idMap[cmd.ID]; ok {
			cmd.LinkLabel = ids
		} else {
			cmd.LinkLabel = []int{}
		}
		cmdClass.Cmds = append(cmdClass.Cmds, cmd)
	}

	ret = &cmdClass
	cmdClass.ID = classID
	err = t.Commit()
	return
}
func updateCmdClass(appid string, classID int, newClassName string) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return
	}
	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := `SELECT count(*) FROM cmd_class WHERE appid = ? AND name = ?`
	row := t.QueryRow(queryStr, appid, newClassName)
	count := 0
	err = row.Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			return
		}
	}

	if count > 0 {
		err = errDuplicate
		return
	}

	queryStr = "UPDATE cmd_class SET name = ? WHERE appid = ? AND id = ?"
	_, err = t.Exec(queryStr, newClassName, appid, classID)
	if err != nil {
		return
	}
	return t.Commit()
}
func addCmdClass(appid string, pid *int, className string) (id int, err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return
	}
	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := `SELECT count(*) FROM cmd_class WHERE appid = ? AND name = ?`
	row := t.QueryRow(queryStr, appid, className)
	count := 0
	err = row.Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			return
		}
	}

	if count > 0 {
		err = errDuplicate
		return
	}

	// parent will always be nil
	queryStr = `INSERT INTO cmd_class (appid, name, parent) VALUES (?, ?, ?)`
	result, err := mySQL.Exec(queryStr, appid, className, pid)
	if err != nil {
		return
	}

	id64, err := result.LastInsertId()
	if err != nil {
		return
	}
	id = int(id64)
	err = t.Commit()
	return
}
func deleteCmdClass(appid string, classID int) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return
	}

	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := "DELETE FROM cmd_class WHERE appid = ? AND id = ?"
	_, err = t.Exec(queryStr, appid, classID)
	if err != nil {
		return
	}

	queryStr = "UPDATE cmd SET cid = NULL WHERE appid = ? AND cid = ?"
	_, err = t.Exec(queryStr, appid, classID)
	if err != nil {
		return
	}

	err = t.Commit()
	return
}

func moveCmd(appid string, id int, cid int) (err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return err
	}

	tx, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(tx)
	pidPtr := &cid
	if cid == -1 {
		pidPtr = nil
	}

	queryStr := `
		UPDATE cmd SET cid = ?
		WHERE cmd_id = ? AND appid = ?`
	_, err = tx.Exec(queryStr, pidPtr, id, appid)
	if err != nil {
		return err
	}

	return tx.Commit()
}



func addCmdImportRecord(appid string, userid string, filename string, rows int, valid_rows int) (int, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return -1, err
	}

	tx, err := mySQL.Begin()
	if err != nil {
		return -1, err
	}
	defer util.ClearTransition(tx)

	queryStr := `
		INSERT INTO tbl_upload_cmd_history
		(user_id, app_id, rows, valid_rows, file_name)
		VALUES (?, ?, ?, ?, ?)`

	queryParams := []interface{}{
		userid,
		appid,
		rows,
		valid_rows,
		filename,
	}
	result, err := tx.Exec(queryStr, queryParams...)
	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}

	return int(id), tx.Commit()
}

func updateCommandImportProgress(recordId int, finish_rows int) (error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return err
	}

	tx, err := mySQL.Begin()
	if err != nil {
		return  err
	}
	defer util.ClearTransition(tx)

	queryStr := `
		UPDATE tbl_upload_cmd_history
        SET finish_rows = ?
        WHERE id = ?`


	queryParams := []interface{}{
		finish_rows,
		recordId,
	}
	_, err = tx.Exec(queryStr, queryParams...)
	if err != nil {
		return  err
	}
	//id, err := result.LastInsertId()
	//if err != nil {
	//	return -1, err
	//}

	return  tx.Commit()
}


func updateCommandReportFile(recordId int, path string) (error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return err
	}

	tx, err := mySQL.Begin()
	if err != nil {
		return  err
	}
	defer util.ClearTransition(tx)

	queryStr := `
		UPDATE tbl_upload_cmd_history
        SET upload_detail_file_path = ?
        WHERE id = ?`


	queryParams := []interface{}{
		path,
		recordId,
	}
	_, err = tx.Exec(queryStr, queryParams...)
	if err != nil {
		return  err
	}
	//id, err := result.LastInsertId()
	//if err != nil {
	//	return -1, err
	//}

	return  tx.Commit()
}
func deleteAllCommand(appid string) (error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return err
	}

	tx, err := mySQL.Begin()
	if err != nil {
		return  err
	}
	defer util.ClearTransition(tx)
	queryStr := `
		DELETE FROM cmd
        WHERE appid = ?`


	queryParams := []interface{}{
		appid,
	}
	_, err = tx.Exec(queryStr, queryParams...)
	if err != nil {
		return  err
	}
	//id, err := result.LastInsertId()
	//if err != nil {
	//	return -1, err
	//}

	return  tx.Commit()
}


func getCommandImportReportPath(recordId int) (string, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return "", errDBNotInit
	}

	queryStr := `
		SELECT upload_detail_file_path
		FROM tbl_upload_cmd_history
		WHERE id = ?`
	row := mySQL.QueryRow(queryStr, recordId)
	path := ""
	err = row.Scan(&path)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	return path, nil
}


func getCommandImportProgress(recordId int) (int, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return 0, errDBNotInit
	}

	queryStr := `
		SELECT rows, finish_rows
		FROM tbl_upload_cmd_history
		WHERE id = ?`
	row := mySQL.QueryRow(queryStr, recordId)
	total := 0
	finish := 0
	err = row.Scan(&total, &finish)

	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if total == 0 {
		return 0, err
	}
	progress := (finish * 100)/total

	return progress, nil
}


func getAllCommand(appid string) (rcommandObjList []*CommandObj, err error) {

	commands, err := combineCommand(appid)
	if err != nil {
		return nil, err
	}

	return commands, nil
}


func combineCommand(appid string) (commandObjList []*CommandObj, err error) {

	commandList, err := getOriginCommand(appid)

	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return
	}

	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	if err == nil {
		for _, command := range commandList {

			queryStr := `SELECT cmd_id, robot_tag_id FROM cmd_robot_tag WHERE cmd_id = ?`
			labelRows, err:= t.Query(queryStr, command.Id)
			if err != nil {
				continue
			}
			defer labelRows.Close()
			cmdTag := CmdTag{}
			for labelRows.Next(){
				labelRows.Scan(&cmdTag.CmdId, &cmdTag.RobotTagId)
				command.Tags = append(command.Tags, cmdTag.RobotTagId)
			}
		}
	}

	err = t.Commit()
	return commandList, err

}


func getOriginCommand(appid string)(commandObjList []*CommandObj, err error){


	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errDBNotInit
		return
	}

	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := `SELECT cmd_id, cid, name, target, rule, answer, response_type, begin_time, end_time FROM cmd WHERE appid = ?`
	cmdRows, err:= t.Query(queryStr, appid)


	for cmdRows.Next() {
		commandObj := CommandObj{}
		err = cmdRows.Scan(&commandObj.Id, &commandObj.ClassId, &commandObj.Name, &commandObj.Target,
			&commandObj.Rule, &commandObj.Answer, &commandObj.ResponseType, &commandObj.BeginTime, &commandObj.EndTime )
		if err != nil {
			continue
		}
		commandObjList = append(commandObjList, &commandObj)

	}

	defer cmdRows.Close()
	err = t.Commit()
	return commandObjList, err

}