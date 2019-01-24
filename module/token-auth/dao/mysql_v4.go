package dao

import (
	"database/sql"

	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
)

// GetOAuthClient will get client info with clientID, if ID is invalid, return nil
func (controller MYSQLController) GetOAuthClient(clientID string) (*data.OAuthClient, error) {
	ok, err := controller.checkDB()
	if !ok {
		util.LogDBError(err)
		return nil, err
	}

	queryStr := `
		SELECT secret, redirect_uri, status
		FROM product
		WHERE id = ?`
	row := controller.connectDB.QueryRow(queryStr, clientID)

	status := 0
	ret := data.OAuthClient{ID: clientID}
	err = row.Scan(&ret.Secret, &ret.RedirectURI, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	ret.Active = status > 0

	return &ret, nil
}
