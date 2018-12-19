package cu

import (
	"database/sql"
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/util"
)

// Dao is the interface of qi dao, it can be used for mock
type Dao interface {
	InitDB() error
	Begin() (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	CreateFlowConversation(tx *sql.Tx, d *daoFlowCreate) (int64, error)
	Group(tx *sql.Tx, query GroupQuery) ([]Group, error)
}

// GroupQuery can used to query the group table
type GroupQuery struct {
	Type         []int
	EnterpriseID *string
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

// Group
type Group struct {
	AppID          uint64
	Name           string
	EnterpriseID   string
	Description    string
	CreatedTime    int64
	UpdatedTime    int64
	IsDelete       bool
	IsEnable       bool
	LimitedSpeed   int
	LimitedSilence float32
	typ            int
}

//SQLDao is sql struct used to access database
type SQLDao struct {
	conn *sql.DB
}

//InitDB is used to get the db in this module
func (s SQLDao) InitDB() error {
	s.conn = GetDB()
	if s.conn == nil {
		return util.ErrDBNotInit
	}
	return nil
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
func (s SQLDao) CreateFlowConversation(tx *sql.Tx, d *daoFlowCreate) (int64, error) {

	if s.conn == nil && tx == nil {
		return 0, util.ErrDBNotInit
	}

	table := Conversation
	insertSQL := fmt.Sprintf("INSERT INTO `%s` (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s) VALUES (?,?,?,?,?,?,?,?,?,?)",
		table,
		ConFieldEnterprise, ConFieldFileName, ConFieldCallTime, ConFieldUpdateTime,
		ConFieldUploadTime, ConFieldType, ConFieldLeftChannel, ConFieldRightChannel,
		ConFieldUUID, ConFieldUser)

	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.Exec(insertSQL, d.enterprise, d.fileName, d.callTime, d.updateTime, d.uploadTime, d.typ, d.leftChannel, d.rightChannel, d.uuid, d.user)
	} else {
		res, err = s.conn.Exec(insertSQL, d.enterprise, d.fileName, d.callTime, d.updateTime, d.uploadTime, d.typ, d.leftChannel, d.rightChannel, d.uuid, d.user)
	}
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

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
		fldGroupDescription + "`, `" + fldGroupCreatedTime + "`, `" + fldGroupUpdatedTime + "`, `" + fldGroupIsDeleted + "`, `" +
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
		var g Group
		rows.Scan(&g.AppID, &g.IsDelete, &g.Name, &g.EnterpriseID, &g.Description, &g.CreatedTime, &g.UpdatedTime, &g.IsDelete, &g.LimitedSpeed, &g.LimitedSilence)
		groups = append(groups, g)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan failed, %v", err)
	}

	return groups, nil
}
