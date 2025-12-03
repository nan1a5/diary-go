package mysql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"diary/config"
)

var globalDB *gorm.DB

func InitDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(cfg.DBDsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	globalDB = db
	return db
}

func CloseDB(db *gorm.DB) {
	// sqlite driver uses underlying sql.DB
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
}
