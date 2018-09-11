package clustering

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

//sqlService is the clustering service implemented by Mysql
type sqlService struct {
	db *sql.DB
}

func (s *sqlService) NewReport(r Report) (uint64, error) {
	query := "INSERT INTO `reports`(`app_id`, `user_id`, `condition`, `created_time`, `updated_time`, `status`, `ignored_size`, `marked_size`) VALUE( ?, ?, ?, ?, ?, ?, ?, ?)"
	result, err := s.db.Exec(query, r.AppID, r.UserID, r.Condition, r.CreatedTime, r.UpdatedTime, r.Status, r.IgnoredSize, r.MarkedSize)
	if err != nil {
		return 0, fmt.Errorf("insert report failed, %v", err)
	}
	id, err := result.LastInsertId()
	//since auto incre id should always be unsighed, it is safe to convert into unsigned number.
	return uint64(id), err
}

func (s *sqlService) GetReport(id uint64) (Report, error) {
	query := "SELECT `id`, `app_id`, `user_id`, `condition`, `created_time`, `updated_time`, `status`, `ignored_size`, `marked_size` FROM `reports` WHERE `id` = ? "
	var r Report
	err := s.db.QueryRow(query, id).Scan(&r.ID, &r.AppID, &r.UserID, &r.Condition, &r.CreatedTime, &r.UpdatedTime, &r.Status, &r.IgnoredSize, &r.MarkedSize)
	if err == sql.ErrNoRows {
		return r, err
	}
	if err != nil {
		return r, fmt.Errorf("sql query row failed, %v", err)
	}
	return r, nil
}

func (s *sqlService) QueryReports(query ReportQuery) ([]Report, error) {
	rawWhere, inputs := asRawSQL(query)
	selectQuery := "SELECT `id`, `app_id`, `user_id`, `condition`, `created_time`, `updated_time`, `status`, `ignored_size`, `marked_size` FROM `reports` WHERE " + rawWhere
	rows, err := s.db.Query(selectQuery, inputs...)
	if err != nil {
		return nil, fmt.Errorf("query sql failed, %v", err)
	}
	defer rows.Close()
	var reports = []Report{}
	for rows.Next() {
		var r Report
		rows.Scan(&r.ID, &r.AppID, &r.UserID, &r.Condition, &r.CreatedTime, &r.UpdatedTime, &r.Status, &r.IgnoredSize, &r.MarkedSize)
		reports = append(reports, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan failed, %v", err)
	}
	return reports, nil
}

func (s *sqlService) NewReportError(rErr ReportError) (uint64, error) {
	query := "INSERT INTO `report_errors` (`report_id`,`cause`, `created_time`) VALUE(?, ?, ?)"
	result, err := s.db.Exec(query, rErr.ReportID, rErr.Cause, rErr.CreateTime)
	if err != nil {
		return 0, fmt.Errorf("sql exec failed, %v", err)
	}
	id, err := result.LastInsertId()
	return uint64(id), err
}

func asRawSQL(query ReportQuery) (string, []interface{}) {
	var preparedInput = []interface{}{}
	var condition = ""
	condition += " `reports`.`app_id` = ? "
	preparedInput = append(preparedInput, query.AppID)
	if query.Reports != nil && len(query.Reports) > 0 {
		condition += "AND `reports`.`id` IN (? " + strings.Repeat(",? ", len(query.Reports)-1) + ")"
		for _, id := range query.Reports {
			preparedInput = append(preparedInput, id)
		}
		return condition, preparedInput
	}
	if query.CreatedTime != nil {
		if query.CreatedTime.StartTime == nil {
			condition += "AND `reports`.`created_time` < ? "
			preparedInput = append(preparedInput, *query.CreatedTime.EndTime)
		} else if query.CreatedTime.EndTime == nil {
			condition += "AND `reports`.`created_time` > ? "
			preparedInput = append(preparedInput, *query.CreatedTime.StartTime)
		} else {
			condition += "AND `reports`.`created_time` BETWEEN ? AND ? "
			preparedInput = append(preparedInput, *query.CreatedTime.StartTime, *query.CreatedTime.EndTime)
		}
	}
	if query.UpdatedTime != nil {
		if query.UpdatedTime.StartTime == nil {
			condition += " AND `reports`.`updated_time` < ? "
			preparedInput = append(preparedInput, *query.UpdatedTime.EndTime)
		} else if query.UpdatedTime.EndTime == nil {
			condition += " AND `reports`.`updated_time` > ? "
			preparedInput = append(preparedInput, *query.UpdatedTime.StartTime)
		} else {
			condition += " AND `reports`.`updated_time` BETWEEN ? AND ? "
			preparedInput = append(preparedInput, *query.UpdatedTime.StartTime, *query.UpdatedTime.EndTime)
		}
	}
	if query.Status != nil {
		condition += " AND `reports`.`status` = ? "
		preparedInput = append(preparedInput, *query.Status)
	}

	if query.UserID != nil {
		condition += " AND `reports`.`user_id` = ? "
		preparedInput = append(preparedInput, *query.UserID)
	}

	return condition, preparedInput
}

func (s *sqlService) UpdateReportStatus(id uint64, status ReportStatus) error {
	query := "UPDATE `reports` SET `status` = ?, `updated_time` = ? WHERE id = ?"
	_, err := s.db.Exec(query, status, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("exec sql failed, %v", err)
	}
	return nil
}

func (s *sqlService) NewCluster(cluster Cluster) (uint64, error) {
	if cluster.ReportID == 0 {
		return 0, fmt.Errorf("report id 0 is invalid")
	}
	query := "INSERT INTO `report_clusters` (`report_id`, `tags`, `created_time`) VALUE(?, ?, ?)"
	result, err := s.db.Exec(query, cluster.ReportID, cluster.Tags, cluster.CreatedTime)
	if err != nil {
		return 0, fmt.Errorf("create cluster from sql failed, %v", err)
	}
	id, err := result.LastInsertId()
	return uint64(id), err
}

func (s *sqlService) GetCluster(id uint64) (Cluster, error) {
	query := "SELECT `id`, `report_id`, `tags`, `created_time` FROM `report_clusters` WHERE id = ?"
	var c Cluster
	err := s.db.QueryRow(query, id).Scan(&c.ID, &c.ReportID, &c.Tags, &c.CreatedTime)
	if err != nil {
		return c, fmt.Errorf("sql query clusters row failed, %v", err)
	}
	return c, nil
}

func (s *sqlService) NewRecords(records ...ReportRecord) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("acquire transcation failed, %v", err)
	}
	insertStmt, err := tx.Prepare("INSERT INTO `report_records`(report_id, cluster_id, chat_record_id, content, is_central_node, created_time) VALUE(?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare insert statement failed, %v", err)
	}
	defer insertStmt.Close()
	for _, r := range records {
		_, err = insertStmt.Exec(r.ReportID, r.ClusterID, r.ChatRecordID, r.Content, r.IsCentralNode, r.CreatedTime)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("exec query failed, %v", err)
		}

	}
	tx.Commit()
	return nil
}

func (s *sqlService) GetRecords(reportID uint64) ([]ReportRecord, error) {
	query := "SELECT `id`, `report_id`, `cluster_id`, `chat_record_id`, `content`, `is_central_node`, `created_time` FROM `report_records` WHERE report_id = ? "
	rows, err := s.db.Query(query, reportID)
	if err != nil {
		return nil, fmt.Errorf("sql query failed, %v", err)
	}
	defer rows.Close()
	var records = []ReportRecord{}
	for rows.Next() {
		var r ReportRecord
		rows.Scan(&r.ID, &r.ReportID, &r.ClusterID, &r.ChatRecordID, &r.Content, &r.IsCentralNode, &r.CreatedTime)
		records = append(records, r)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan failed, %v", err)
	}

	return records, nil
}

func (s *sqlService) GetFTModel(appID string) (string, error) {
	query := "SELECT `value` FROM `ent_config_appid_customization` WHERE `name` LIKE 'ssm_config' AND `app_id` LIKE ?"
	var value string
	err := s.db.QueryRow(query, appID).Scan(&value)
	if err == sql.ErrNoRows {
		return "", err
	}
	if err != nil {
		return "", fmt.Errorf("sql query row failed, %v", err)
	}
	var config ssmConfig
	err = json.Unmarshal([]byte(value), &config)
	if err != nil {
		return "", fmt.Errorf("unmarshal value failed, %v", err)
	}
	var elements []ssmValueElement
	for _, item := range config.Items {
		if item.Name != "dependency" {
			continue
		}

		err = json.Unmarshal(*item.Value, &elements)
		if err != nil {
			return "", fmt.Errorf("unmarshal ssmValueElement failed, %v", err)
		}
	}
	var p ssmParameters
	for _, e := range elements {
		if e.Name != "ml" {
			continue
		}
		err = json.Unmarshal(*e.Parameters, &p)
		if err != nil {
			return "", fmt.Errorf("unmarshal ssmParameters failed, %v", err)
		}
	}

	return p.Model, nil
}
