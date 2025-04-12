package storage

import "database/sql"

type Storage struct {
	DB             *sql.DB
	UserStorage    UserStorage
	ProductStorage ProductStorage
	PVZStorage     PVZStorageI
}
