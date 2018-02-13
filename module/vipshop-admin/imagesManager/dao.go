package imagesManager

import (
	"database/sql"
	"fmt"
)

var mainDB *sql.DB

//Save copy image struct into SQL Database, return nil if success
func Save(image Image) (int64, error) {
	if image.Location == "" {
		return 0, fmt.Errorf("image location should not be empty")
	}
	if image.ID == 0 {
		return 0, fmt.Errorf("image ID should not be zero")
	}
	if image.FileName == "" {
		return 0, fmt.Errorf("image FileName should not be empty")
	}

	result, err := mainDB.Exec("INSERT INTO (id, filename, location, createdTime, lastmodified) VALUES (?, ?, ?, ?, ?)", image.ID, image.FileName, image.Location, image.CreatedTime, image.LastModifiedTime)
	if err != nil {
		return 0, fmt.Errorf("sql exec failed, %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return id, fmt.Errorf("get insert id failed, %v", err)
	}

	return id, nil

}
