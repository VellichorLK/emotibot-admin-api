package handlers

import (
	"database/sql"
	"strings"
)

func getEmotionData(selectColumns [2][]string, whereState [2][]WhereStates, groupIDs []interface{}, orderState [2][]string, params []interface{}) (*sql.Rows, error) {

	sql := getEmotionBaseSQL(selectColumns, whereState, groupIDs, orderState)

	if groupIDs != nil && len(groupIDs) > 0 {
		params = append(params, groupIDs...)
	}

	rows, err := db.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func getEmotionBaseSQL(selectColumns [2][]string, whereState [2][]WhereStates, groupIDs []interface{}, orderState [2][]string) string {

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

	var hasWhereState bool
	for idx, states := range whereState {
		for idx2, state := range states {
			if idx == 0 && idx2 == 0 {
				sql += " where "
				hasWhereState = true
			} else {
				sql += " and "
			}
			sql += nickTable[idx] + "." + state.name + state.compare + "?"
		}
	}

	if groupIDs != nil && len(groupIDs) > 0 {
		if !hasWhereState {
			sql += " where "
		} else {
			sql += " and "
		}
		sql += nickTable[0] + "." + NTAG2 + " in (?" + strings.Repeat(",?", len(groupIDs)-1) + ")"

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
