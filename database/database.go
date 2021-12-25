package database

import (
	"fmt"

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

	DBConn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
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
