package clustering

import "database/sql"

//sqlService is the clustering service implemented by mysql
type sqlService struct {
	db *sql.DB
}
