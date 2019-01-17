package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/util"
)

// Dao is the interface of qi dao, it can be used for mock
type RuleDao interface {
	Begin() (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	CreateFlowConversation(tx *sql.Tx, d *FlowCreate) (int64, error)
	InsertSegment(tx *sql.Tx, seg *Segment) (int64, error)
	GetConversationByUUID(tx *sql.Tx, uuid string) (*ConversationInfo, error)
	GetSegmentByCallID(tx *sql.Tx, callID uint64) ([]*Segment, error)
	Group(tx *sql.Tx, query GroupQuery) ([]Group, error)
	GetGroupToLogicID(tx *sql.Tx, appID uint64) (map[uint64][]uint64, []uint64, error)
	GetRule(tx *sql.Tx, query RuleQuery) ([]*Rule, error)
	GetLogic(tx *sql.Tx, query LogicQuery) ([]*Logic, error)
	InsertFlowResultTmp(tx *sql.Tx, callID uint64, val string) (int64, error)
	UpdateFlowResultTmp(tx *sql.Tx, callID uint64, val string) (int64, error)
	GetFlowResultFromTmp(tx *sql.Tx, callID uint64) (*QIFlowResult, error)
	UpdateConversation(tx *sql.Tx, callID uint64, params map[string]interface{}) (int64, error)
	GetRecommendations(tx *sql.Tx, logicIDs []uint64) (map[uint64][]string, error)
}

type FlowCreate struct {
	Typ          int
	LeftChannel  int
	RightChannel int
	Enterprise   string
	CallTime     int64
	UploadTime   int64
	UpdateTime   int64
	FileName     string
	UUID         string
	User         string
}

//Segment is vad segment
type Segment struct {
	CallID    uint64
	ASR       *AsrContent
	Channel   int
	CreatTime int64
}

type AsrContent struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Text      string  `json:"text"`
	Speaker   string  `json:"speaker"`
}

//ConversationInfo is information in Conversation table
type ConversationInfo struct {
	CallID       uint64
	Status       int
	FileName     string
	FilePath     string
	VoiceID      uint64
	CallComment  string
	Transaction  int
	Series       string
	CallTime     int64
	UploadTime   int64
	UpdateTime   int64
	HostID       string
	HostName     string
	Extension    string
	Department   string
	GuestID      string
	GuestName    string
	GuestPhone   string
	CallUUID     string
	Enterprise   string
	User         string
	Duration     int
	ApplyGroup   []uint64
	Type         int
	LeftChannel  int
	RightChannel int
}

// GroupQuery can used to query the group table
type GroupQuery struct {
	Type         []int
	EnterpriseID *string
}

//RuleQuery gives the query condition for Rule table
type RuleQuery struct {
	ID           []uint64
	EnterpriseID *string
}

//LogicQuery gives the query condition for Logic table
type LogicQuery struct {
	ID           []uint64
	EnterpriseID *string
}

//QIFlowResult give the reuslt of qi flow
type QIFlowResult struct {
	FileName  string               `json:"file_name"`
	Result    []*QIFlowGroupResult `json:"cu_result"`
	Sensitive []string             `json:"sensitive"`
}

//QIFlowGroupResult gives the result of check
type QIFlowGroupResult struct {
	ID       uint64      `json:"-"`
	Name     string      `json:"group_name"`
	QIResult []*QIResult `json:"qi_result"`
}

//QIResult gives the result of rule
type QIResult struct {
	ID          uint64         `json:"-"`
	Name        string         `json:"controller_rule"`
	Valid       bool           `json:"valid"`
	LogicResult []*LogicResult `json:"logic_results"`
}

//LogicResult give the result of logic
type LogicResult struct {
	ID        uint64   `json:"-"`
	Name      string   `json:"logic_rule"`
	Valid     bool     `json:"valid"`
	Recommend []string `json:"recommend"`
}

func (g *GroupQuery) whereSQL() (whereSQL string, bindData []interface{}) {
	bindData = make([]interface{}, 0, 2)
	whereSQL = "WHERE "
	conditions := []string{}
	if g.Type != nil || len(g.Type) > 0 {
		condition := fldGroupType + " IN (?" + strings.Repeat(",?", len(g.Type)-1) + ")"
		conditions = append(conditions, condition)
		for _, t := range g.Type {
			bindData = append(bindData, t)
		}
	}
	if g.EnterpriseID != nil {
		condition := fldGroupEnterprise + " = ?"
		conditions = append(conditions, condition)
		bindData = append(bindData, *g.EnterpriseID)
	}
	whereSQL += strings.Join(conditions, " AND ")
	return whereSQL, bindData
}

func (r *RuleQuery) whereSQL() (whereSQL string, bindData []interface{}) {
	bindData = make([]interface{}, 0, 2)
	whereSQL = "WHERE "
	conditions := []string{}

	if len(r.ID) > 0 {
		condition := fldRuleID + " IN (?" + strings.Repeat(",?", len(r.ID)-1) + ")"
		conditions = append(conditions, condition)
		for _, id := range r.ID {
			bindData = append(bindData, id)
		}
	}

	if r.EnterpriseID != nil {
		condition := fldRuleEnterprise + " = ?"
		conditions = append(conditions, condition)
		bindData = append(bindData, *r.EnterpriseID)
	}
	whereSQL += strings.Join(conditions, " AND ")
	return whereSQL, bindData
}

func (l *LogicQuery) whereSQL() (whereSQL string, bindData []interface{}) {
	bindData = make([]interface{}, 0, 2)
	whereSQL = "WHERE "
	conditions := []string{}

	if len(l.ID) > 0 {
		condition := fldLogicID + " IN (?" + strings.Repeat(",?", len(l.ID)-1) + ")"
		conditions = append(conditions, condition)
		for _, id := range l.ID {
			bindData = append(bindData, id)
		}
	}

	if l.EnterpriseID != nil {
		condition := fldLogicEnterprise + " = ?"
		conditions = append(conditions, condition)
		bindData = append(bindData, *l.EnterpriseID)
	}
	whereSQL += strings.Join(conditions, " AND ")
	return whereSQL, bindData
}

//Rule is field in Rule
type Rule struct {
	RuleID      uint64
	IsDelete    int
	Name        string
	Method      int
	Score       int
	Description string
	Enterprise  string
}

//Logic is field in Logic
type Logic struct {
	LogicID         uint64
	Name            string
	TagDistance     int
	RangeConstraint string
	CreateTime      int64
	UpdateTime      int64
	IsDelete        int
	Enterprise      string
	Speaker         int
}

//SQLDao is sql struct used to access database
type SQLDao struct {
	conn *sql.DB
}

func NewSQLDao(conn *sql.DB) *SQLDao {
	return &SQLDao{
		conn: conn,
	}
}

//Begin is used to start a transaction
func (s SQLDao) Begin() (*sql.Tx, error) {
	if s.conn == nil {
		return nil, util.ErrDBNotInit
	}
	return s.conn.Begin()

}

//Commit commits the data
func (s SQLDao) Commit(tx *sql.Tx) error {
	if tx != nil {
		return tx.Commit()
	}
	return nil
}

//CreateFlowConversation creates the flow conversation
func (s SQLDao) CreateFlowConversation(tx *sql.Tx, d *FlowCreate) (int64, error) {

	if s.conn == nil && tx == nil {
		return 0, util.ErrDBNotInit
	}

	table := tblConversation
	insertSQL := fmt.Sprintf("INSERT INTO `%s` (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s) VALUES (?,?,?,?,?,?,?,?,?,?)",
		table,
		ConFieldEnterprise, ConFieldFileName, ConFieldCallTime, ConFieldUpdateTime,
		ConFieldUploadTime, ConFieldType, ConFieldLeftChannel, ConFieldRightChannel,
		ConFieldUUID, ConFieldUser)

	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.Exec(insertSQL, d.Enterprise, d.FileName, d.CallTime, d.UpdateTime, d.UploadTime, d.Typ, d.LeftChannel, d.RightChannel, d.UUID, d.User)
	} else {
		res, err = s.conn.Exec(insertSQL, d.Enterprise, d.FileName, d.CallTime, d.UpdateTime, d.UploadTime, d.Typ, d.LeftChannel, d.RightChannel, d.UUID, d.User)
	}
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

//InsertSegment inserts the segment
func (s SQLDao) InsertSegment(tx *sql.Tx, seg *Segment) (int64, error) {
	if s.conn == nil && tx == nil {
		return 0, util.ErrDBNotInit
	}

	table := tblSegment
	insertSQL := fmt.Sprintf("INSERT INTO `%s` (%s,%s,%s,%s,%s,%s) VALUES (?,?,?,?,?,?)",
		table,
		fldSegmentCallID, fldSegmentStartTime, fldSegmentEndTime, fldSegmentChannel, fldSegmentCreateTime, fldSegmentText,
	)

	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.Exec(insertSQL, seg.CallID, seg.ASR.StartTime, seg.ASR.EndTime, seg.Channel, seg.CreatTime, seg.ASR.Text)
	} else {
		res, err = s.conn.Exec(insertSQL, seg.CallID, seg.ASR.StartTime, seg.ASR.EndTime, seg.Channel, seg.CreatTime, seg.ASR.Text)
	}
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

//GetConversationByUUID gets the conversation information
func (s SQLDao) GetConversationByUUID(tx *sql.Tx, uuid string) (*ConversationInfo, error) {
	if s.conn == nil && tx == nil {
		return nil, util.ErrDBNotInit
	}

	table := tblConversation

	querySQL := fmt.Sprintf("SELECT %s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s FROM %s WHERE %s=?",
		ConFieldID, ConFieldStatus, ConFieldFileName, ConFieldPath, ConFieldVoiceID,
		ConFieldCallComment, ConFieldTransaction, ConFieldSeries, ConFieldCallTime, ConFieldUploadTime,
		ConFieldUpdateTime, ConFieldHostID, ConFieldHostName, ConFieldExtenstion, ConFieldDepartment,
		ConFieldGuestID, ConFieldGuestName, ConFieldGuestPhone, ConFieldUUID, ConFieldEnterprise,
		ConFieldUser, ConFieldDuration, ConFieldApplyGroup, ConFieldType, ConFieldLeftChannel,
		ConFieldRightChannel,
		table, ConFieldUUID)
	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.Query(querySQL, uuid)
	} else {
		rows, err = s.conn.Query(querySQL, uuid)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var callID uint64
	var callTime, uploadTime, updateTime int64
	var status, transaction, duration, typ, leftChannel, rightChannel int
	var fileName, filePath, callComment, applyGroup sql.NullString
	var series, hostID, hostName, extension, department, guestID, guestName, guestPhone, calluuid, enterprise, user string
	var voiceID sql.NullInt64

	if rows.Next() {
		err = rows.Scan(&callID, &status, &fileName, &filePath, &voiceID,
			&callComment, &transaction, &series, &callTime, &uploadTime,
			&updateTime, &hostID, &hostName, &extension, &department,
			&guestID, &guestName, &guestPhone, &calluuid, &enterprise,
			&user, &duration, &applyGroup, &typ, &leftChannel,
			&rightChannel)
		if err != nil {
			return nil, err
		}

		info := &ConversationInfo{CallID: callID, Status: status, FileName: fileName.String, FilePath: filePath.String, VoiceID: uint64(voiceID.Int64),
			CallComment: callComment.String, Transaction: transaction, Series: series, CallTime: callTime, UploadTime: uploadTime,
			UpdateTime: updateTime, HostID: hostID, HostName: hostName, Extension: extension, Department: department,
			GuestID: guestID, GuestName: guestName, GuestPhone: guestPhone, CallUUID: calluuid, Enterprise: enterprise,
			User: user, Duration: duration, Type: typ, LeftChannel: leftChannel, RightChannel: rightChannel}

		if applyGroup.Valid {
			var group []uint64
			if err := json.Unmarshal([]byte(applyGroup.String), &group); err != nil {
				return nil, err
			}
			info.ApplyGroup = group
		}

		return info, nil

	}
	return nil, nil

}

//GetSegmentByCallID gets the segments from Segment table
func (s SQLDao) GetSegmentByCallID(tx *sql.Tx, callID uint64) ([]*Segment, error) {
	if s.conn == nil && tx == nil {
		return nil, util.ErrDBNotInit
	}

	table := tblSegment
	querySQL := fmt.Sprintf("SELECT %s,%s,%s,%s,%s,%s FROM %s WHERE %s=? ORDER BY %s ASC",
		fldSegmentID, fldSegmentStartTime, fldSegmentEndTime, fldSegmentChannel,
		fldSegmentCreateTime, fldSegmentText,
		table, fldSegmentCallID, fldSegmentStartTime)
	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.Query(querySQL, callID)
	} else {
		rows, err = s.conn.Query(querySQL, callID)
	}
	defer rows.Close()

	segments := make([]*Segment, 0, 0)
	var id uint64
	var startT, endT float64
	var channel int
	var createTime int64
	var asrText string
	for rows.Next() {
		err = rows.Scan(&id, &startT, &endT, &channel, &createTime, &asrText)
		if err != nil {
			return nil, err
		}
		asr := &AsrContent{StartTime: startT, EndTime: endT, Text: asrText}
		segment := &Segment{CallID: callID, Channel: channel, CreatTime: createTime, ASR: asr}
		segments = append(segments, segment)
	}
	return segments, nil
}

//TODO Reractor this to group dao
func (s SQLDao) Group(tx *sql.Tx, query GroupQuery) ([]Group, error) {
	type queryer interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
	}
	var q queryer
	if tx != nil {
		q = tx
	} else if s.conn != nil {
		q = s.conn
	} else {
		return nil, util.ErrDBNotInit
	}

	sqlQuery := "SELECT `" + fldGroupAppID + "`, `" + fldGroupIsDeleted + "`, `" + fldGroupName + "`, `" + fldGroupEnterprise + "`, `" +
		fldGroupDescription + "`, `" + fldGroupCreatedTime + "`, `" + fldGroupUpdatedTime + "`, `" + fldGroupIsEnabled + "`, `" +
		fldGroupLimitedSpeed + "`, `" + fldGroupLimitedSilence + "`, `" + fldGroupType + "` FROM `" + tblGroup + "`"
	wherePart, bindData := query.whereSQL()
	if len(bindData) > 0 {
		sqlQuery += " " + wherePart
	}
	rows, err := q.Query(sqlQuery, bindData...)
	if err != nil {
		logger.Error.Println("raw sql: ", sqlQuery)
		logger.Error.Println("raw bind-data: ", bindData)
		return nil, fmt.Errorf("sql executed failed, %v", err)
	}
	defer rows.Close()
	var groups = make([]Group, 0)
	for rows.Next() {

		var g = Group{}
		var isDeleted, isEnabled int
		rows.Scan(&g.AppID, &isDeleted, &g.Name, &g.EnterpriseID, &g.Description, &g.CreatedTime, &g.UpdatedTime, &isEnabled, &g.LimitedSpeed, &g.LimitedSilence, &g.typ)
		if isDeleted == 1 {
			g.IsDelete = true
		}
		if isEnabled == 1 {
			g.IsEnable = true
		}
		groups = append(groups, g)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("sql scan failed, %v", err)
	}

	return groups, nil
}

//GetGroupToLogicID gets the id and name of logics in rule, rules in group
func (s SQLDao) GetGroupToLogicID(tx *sql.Tx, groupID uint64) (map[uint64][]uint64, []uint64, error) {
	type queryer interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
	}
	var q queryer
	if tx != nil {
		q = tx
	} else if s.conn != nil {
		q = s.conn
	} else {
		return nil, nil, util.ErrDBNotInit
	}

	tableA := "Relation_Group_Rule"
	tableB := tblRelRuleLogic
	findIDSQL := fmt.Sprintf("SELECT a.%s,b.%s FROM %s AS a LEFT JOIN %s AS b on a.%s=b.%s WHERE a.%s=?",
		fldRuleID, fldLogicID,
		tableA, tableB,
		fldRuleID, fldRuleID,
		fldGroupAppID)
	rows, err := q.Query(findIDSQL, groupID)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		return nil, nil, err
	}
	defer rows.Close()
	var ruleID uint64
	var logicID sql.NullInt64

	ruleIDToLogicID := make(map[uint64][]uint64)
	ruleOrder := make([]uint64, 0)

	for rows.Next() {
		err = rows.Scan(&ruleID, &logicID)
		if err != nil {
			logger.Error.Printf("%s\n", err)
			return nil, nil, err
		}

		if _, ok := ruleIDToLogicID[ruleID]; !ok {
			ruleIDToLogicID[ruleID] = make([]uint64, 0)
			ruleOrder = append(ruleOrder, ruleID)
		}

		if logicID.Valid {
			ruleIDToLogicID[ruleID] = append(ruleIDToLogicID[ruleID], uint64(logicID.Int64))
		}
	}
	return ruleIDToLogicID, ruleOrder, nil
}

//GetRule gets rule information based on condition
func (s SQLDao) GetRule(tx *sql.Tx, query RuleQuery) ([]*Rule, error) {
	type queryer interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
	}
	var q queryer
	if tx != nil {
		q = tx
	} else if s.conn != nil {
		q = s.conn
	} else {
		return nil, util.ErrDBNotInit
	}

	sqlQuery := fmt.Sprintf("SELECT `%s`,`%s`,`%s`,`%s`,`%s`,`%s`,`%s` FROM `%s`",
		fldRuleID, fldRuleIsDelete, fldRuleName, fldRuleMethod, fldRuleScore, fldRuleDescription, fldRuleEnterprise,
		tblRule)

	wherePart, bindData := query.whereSQL()
	if len(bindData) > 0 {
		sqlQuery += " " + wherePart
	}
	rows, err := q.Query(sqlQuery, bindData...)
	if err != nil {
		logger.Error.Println("raw sql: ", sqlQuery)
		logger.Error.Println("raw bind-data: ", bindData)
		return nil, fmt.Errorf("sql executed failed, %v", err)
	}
	defer rows.Close()
	rules := make([]*Rule, 0)
	for rows.Next() {
		var r Rule
		err = rows.Scan(&r.RuleID, &r.IsDelete, &r.Name, &r.Method, &r.Score, &r.Description, &r.Enterprise)
		if err != nil {
			logger.Error.Printf("%s\n", err)
			return nil, err
		}
		rules = append(rules, &r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan failed, %v", err)
	}

	return rules, nil
}

//GetLogic gets logic information based on condition
func (s SQLDao) GetLogic(tx *sql.Tx, query LogicQuery) ([]*Logic, error) {
	type queryer interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
	}
	var q queryer
	if tx != nil {
		q = tx
	} else if s.conn != nil {
		q = s.conn
	} else {
		return nil, util.ErrDBNotInit
	}

	sqlQuery := fmt.Sprintf("SELECT `%s`,`%s`,`%s`,`%s`,`%s`,`%s`,`%s`,`%s`,`%s` FROM `%s`",
		fldLogicID, fldLogicName, fldLogicTagDist, fldLogicRangeConstraint, fldLogicCreateTime,
		fldLogicUpdateTime, fldLogicIsDelete, fldLogicEnterprise, fldLogicSpeaker,
		tblLogic)

	wherePart, bindData := query.whereSQL()
	if len(bindData) > 0 {
		sqlQuery += " " + wherePart
	}
	rows, err := q.Query(sqlQuery, bindData...)
	if err != nil {
		logger.Error.Println("raw sql: ", sqlQuery)
		logger.Error.Println("raw bind-data: ", bindData)
		return nil, fmt.Errorf("sql executed failed, %v", err)
	}
	defer rows.Close()
	logics := make([]*Logic, 0)
	for rows.Next() {
		var l Logic
		var rangeConstraint *sql.NullString
		err = rows.Scan(&l.LogicID, &l.Name, &l.TagDistance, &rangeConstraint, &l.CreateTime,
			&l.UpdateTime, &l.IsDelete, &l.Enterprise, &l.Speaker)
		if err != nil {
			logger.Error.Printf("%s\n", err)
			return nil, err
		}
		if rangeConstraint.Valid {
			l.RangeConstraint = rangeConstraint.String
		}
		logics = append(logics, &l)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan failed, %v", err)
	}

	return logics, nil
}

//InsertFlowResultTmp inserts the record the CUPredict for now
func (s SQLDao) InsertFlowResultTmp(tx *sql.Tx, callID uint64, val string) (int64, error) {
	type executor interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
	}
	var q executor
	if tx != nil {
		q = tx
	} else if s.conn != nil {
		q = s.conn
	} else {
		return 0, util.ErrDBNotInit
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (%s,%s,%s) VALUES (?,0,?)", tblCUPredict, fldCallID, fldGroupAppID, fldCUPredict)
	result, err := q.Exec(insertSQL, callID, val)
	if err != nil {
		logger.Error.Println("raw sql: ", insertSQL)
		logger.Error.Printf("raw bind-data: [%v,%v]\n", callID, val)
		return 0, fmt.Errorf("sql executed failed, %v", err)
	}
	return result.LastInsertId()
}

//UpdateFlowResultTmp update the record the CUPredict for now
func (s SQLDao) UpdateFlowResultTmp(tx *sql.Tx, callID uint64, val string) (int64, error) {
	type executor interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
	}
	var q executor
	if tx != nil {
		q = tx
	} else if s.conn != nil {
		q = s.conn
	} else {
		return 0, util.ErrDBNotInit
	}

	updateSQL := fmt.Sprintf("UPDATE %s SET %s=? WHERE %s=?", tblCUPredict, fldCUPredict, fldCallID)
	result, err := q.Exec(updateSQL, val, callID)
	if err != nil {
		logger.Error.Println("raw sql: ", updateSQL)
		logger.Error.Printf("raw bind-data: [%v,%v]\n", callID, val)
		return 0, fmt.Errorf("sql executed failed, %v", err)
	}
	return result.RowsAffected()
}

//GetFlowResultFromTmp gets the flow from tmp
func (s SQLDao) GetFlowResultFromTmp(tx *sql.Tx, callID uint64) (*QIFlowResult, error) {
	type queryer interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
	}
	var q queryer
	if tx != nil {
		q = tx
	} else if s.conn != nil {
		q = s.conn
	} else {
		return nil, util.ErrDBNotInit
	}

	querySQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s=?", fldCUPredict, tblCUPredict, fldCallID)
	rows, err := q.Query(querySQL, callID)
	if err != nil {
		logger.Error.Println("raw sql: ", querySQL)
		logger.Error.Printf("raw bind-data: [%v]\n", callID)
		return nil, fmt.Errorf("sql executed failed, %v", err)
	}

	var val *sql.NullString
	result := &QIFlowResult{Result: make([]*QIFlowGroupResult, 0), Sensitive: make([]string, 0)}
	if rows.Next() {
		err = rows.Scan(&val)
		if err != nil {
			logger.Error.Printf("Scan %s failed. %s\n", fldCUPredict, err)
			return nil, err
		}
		if val.Valid {
			err = json.Unmarshal([]byte(val.String), result)
			if err != nil {
				logger.Error.Printf("Marshal json failed. %s\n", err)
				return nil, err
			}
		}
	}
	return result, nil
}

//UpdateConversation updates the info in Conversation table
func (s SQLDao) UpdateConversation(tx *sql.Tx, callID uint64, params map[string]interface{}) (int64, error) {
	type executor interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
	}
	var q executor
	if tx != nil {
		q = tx
	} else if s.conn != nil {
		q = s.conn
	} else {
		return 0, util.ErrDBNotInit
	}

	numOfParams := len(params)
	bindData := make([]interface{}, 0, numOfParams+1)
	setFields := make([]string, 0, numOfParams)
	for k, v := range params {
		field := k + "=?"
		setFields = append(setFields, field)
		bindData = append(bindData, v)
	}

	setStat := strings.Join(setFields, ",")
	bindData = append(bindData, callID)
	updateSQL := fmt.Sprintf("UPDATE %s SET %s WHERE %s=?", tblConversation, setStat, fldCallID)

	result, err := q.Exec(updateSQL, bindData...)
	if err != nil {
		logger.Error.Println("raw sql: ", updateSQL)
		logger.Error.Printf("raw bind-data: [%v]\n", callID)
		return 0, fmt.Errorf("sql executed failed, %v", err)
	}
	return result.RowsAffected()
}

//GetRecommendations gets the recommendation wordings for each logic
func (s SQLDao) GetRecommendations(tx *sql.Tx, logicIDs []uint64) (map[uint64][]string, error) {
	if len(logicIDs) == 0 {
		return nil, nil
	}

	type queryer interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
	}
	var q queryer
	if tx != nil {
		q = tx
	} else if s.conn != nil {
		q = s.conn
	} else {
		return nil, util.ErrDBNotInit
	}

	querySQL := fmt.Sprintf("SELECT %s,%s FROM %s WHERE %s in (?%s)", fldLinkID, fldSentence, tblRecommend, fldLinkID, strings.Repeat(",?", len(logicIDs)-1))

	numOfLogicID := len(logicIDs)
	params := make([]interface{}, numOfLogicID, numOfLogicID)

	for i := 0; i < numOfLogicID; i++ {
		params[i] = logicIDs[i]
	}
	rows, err := q.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query sql:%s, params:%+v failed. %s\n", querySQL, params, err)
		return nil, err
	}
	defer rows.Close()

	logicRecommendMap := make(map[uint64][]string)
	var logicID uint64
	var sentence string
	var sentences []string
	var ok bool
	for rows.Next() {
		err = rows.Scan(&logicID, &sentence)
		if err != nil {
			logger.Error.Printf("Scan error. %s\n", err)
			return nil, err
		}
		if sentences, ok = logicRecommendMap[logicID]; !ok {
			sentences = make([]string, 0)
		}
		sentences = append(sentences, sentence)
		logicRecommendMap[logicID] = sentences
	}

	return logicRecommendMap, nil
}
