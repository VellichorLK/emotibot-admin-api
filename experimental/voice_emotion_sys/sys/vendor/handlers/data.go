package handlers

import (
	"database/sql"
)

func getEmotionData(selectColumns [2][]string, whereState [2][]WhereStates, orderState [2][]string, params []interface{}) (*sql.Rows, error) {

	sql := getEmotionBaseSQL(selectColumns, whereState, orderState)

	rows, err := db.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func getEmotionBaseSQL(selectColumns [2][]string, whereState [2][]WhereStates, orderState [2][]string) string {

	nickTable := [2]string{"a", "b"}

	sql := "select "

	for idx, columns := range selectColumns {
		for idx2, column := range columns {
			if idx == 0 && idx2 == 0 {
			} else {
				sql += ","
			}
			sql += nickTable[idx] + "." + column
		}
	}

	sql += " from " + MainTable + " as " + nickTable[0] + " inner join " + ChannelTable + " as " + nickTable[1] +
		" on " + nickTable[0] + "." + NID + "=" + nickTable[1] + "." + NID

	for idx, states := range whereState {
		for idx2, state := range states {
			if idx == 0 && idx2 == 0 {
				sql += " where "
			} else {
				sql += " and "
			}
			sql += nickTable[idx] + "." + state.name + state.compare + "?"
		}
	}

	for idx, orders := range orderState {
		for idx2, order := range orders {
			if idx == 0 && idx2 == 0 {
				sql += " order by "
			} else {
				sql += ","
			}

			sql += nickTable[idx] + "." + order
		}
	}

	return sql
}
