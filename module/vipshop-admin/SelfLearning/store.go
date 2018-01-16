package SelfLearning

type dbStore struct{}

func (store *dbStore) Store(cr *clusteringResult) error {

	tx, err := getTx()
	if err != nil {
		return err
	}
	defer tx.Commit()

	resultSQL := "insert into " + TableProps.clusterResult.name +
		" (" + TableProps.clusterResult.feedbackID + "," + TableProps.clusterResult.reportID + "," + TableProps.clusterResult.clusterID + ") " +
		" values(?,?,?)"

	tagSQL := "insert into " + TableProps.clusterTag.name +
		" (" + TableProps.clusterTag.reportID + "," + TableProps.clusterTag.clusteringID + "," + TableProps.clusterTag.tag + ") values (?,?,?)"

	resultStmt, err := tx.Prepare(resultSQL)
	if err != nil {
		return err
	}
	defer resultStmt.Close()

	tagStmt, err := tx.Prepare(tagSQL)
	if err != nil {
		return err
	}
	defer tagStmt.Close()

	for idx, cluster := range cr.clusters {
		for _, qID := range cluster.feedbackID {
			_, err := resultStmt.Exec(qID, cr.reportID, idx)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		for _, tag := range cluster.tags {
			_, err := tagStmt.Exec(cr.reportID, idx, tag)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

	}

	return nil
}
