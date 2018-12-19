package cu

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"emotibot.com/emotigo/module/admin-api/util"
)

// Dao is the interface of qi dao, it can be used for mock
type Dao interface {
	InitDB() error
	Begin() (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	CreateFlowConversation(tx *sql.Tx, d *daoFlowCreate) (int64, error)
	InsertSegment(tx *sql.Tx, seg *Segment) (int64, error)
	GetConversationByUUID(tx *sql.Tx, uuid string) (*ConversationInfo, error)
	GetSegmentByCallID(tx *sql.Tx, callID uint64) ([]*Segment, error)
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

//InsertSegment inserts the segment
func (s SQLDao) InsertSegment(tx *sql.Tx, seg *Segment) (int64, error) {
	if s.conn == nil && tx == nil {
		return 0, util.ErrDBNotInit
	}

	table := TableSegment
	insertSQL := fmt.Sprintf("INSERT INTO `%s` (%s,%s,%s,%s,%s,%s) VALUES (?,?,?,?,?,?)",
		table,
		SegFieldCallID, SegFieldStartTime, SegFieldEndTime, SegFieldChannel, SegFieldCreateTiem, SegFieldAsrText,
	)

	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.Exec(insertSQL, seg.callID, seg.asr.StartTime, seg.asr.EndTime, seg.channel, seg.creatTime, seg.asr.Text)
	} else {
		res, err = s.conn.Exec(insertSQL, seg.callID, seg.asr.StartTime, seg.asr.EndTime, seg.channel, seg.creatTime, seg.asr.Text)
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

	table := Conversation

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

	table := TableSegment
	querySQL := fmt.Sprintf("SELECT %s,%s,%s,%s,%s,%s FROM %s WHERE %s=? ORDER BY %s ASC",
		SegFieldID, SegFieldStartTime, SegFieldEndTime, SegFieldChannel,
		SegFieldCreateTiem, SegFieldAsrText,
		table, SegFieldCallID, SegFieldStartTime)
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
		asr := &apiFlowAddBody{StartTime: startT, EndTime: endT, Text: asrText}
		segment := &Segment{callID: callID, channel: channel, creatTime: createTime, asr: asr}
		segments = append(segments, segment)
	}
	return segments, nil
}
