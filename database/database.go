package database

import (
	"fmt"
	"time"

	"github.com/mrrizal/devcode-backend-challenge/configs"
	"github.com/mrrizal/devcode-backend-challenge/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DBConn *gorm.DB

func InitDatabase(config configs.Conf) error {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.MysqlUser,
		config.MysqlPassword, config.MysqlHost, config.MysqlDBName)
	DBConn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})

	db, _ := DBConn.DB()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(time.Duration(5 * time.Minute.Minutes()))

	if err != nil {
		return err
	}

	if err := DBConn.AutoMigrate(&models.TodoModel{}); err != nil {
		return err
	}

	if err := DBConn.AutoMigrate(&models.ActivityModel{}); err != nil {
		return err
	}

	return nil
}
