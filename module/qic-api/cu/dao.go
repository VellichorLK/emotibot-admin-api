package cu

import (
	"database/sql"
	"fmt"

	"emotibot.com/emotigo/module/admin-api/util"
)

// Dao is the interface of qi dao, it can be used for mock
type Dao interface {
	InitDB() error
	Begin() (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	CreateFlowConversation(tx *sql.Tx, d *daoFlowCreate) (int64, error)
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
