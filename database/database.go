package database

import (
	"fmt"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mrrizal/devcode-backend-challenge/configs"
)

var DBConn *sql.DB

func autoMigrate() error {
	activitiesTAble := `
	CREATE TABLE IF NOT EXISTS activities (
		id bigint NOT NULL AUTO_INCREMENT,
		email longtext,
		title longtext,
		created_at datetime(3) DEFAULT NULL,
		updated_at datetime(3) DEFAULT NULL,
		deleted_at datetime(3) DEFAULT NULL,
		PRIMARY KEY (id),
		KEY idx_activities_deleted_at (deleted_at)
	  )

	`

	todosTable := `
	CREATE TABLE IF NOT EXISTS todos (
		id bigint NOT NULL AUTO_INCREMENT,
		created_at datetime(3) DEFAULT NULL,
		updated_at datetime(3) DEFAULT NULL,
		deleted_at datetime(3) DEFAULT NULL,
		activity_group_id bigint DEFAULT NULL,
		title longtext,
		is_active longtext,
		priority longtext,
		PRIMARY KEY (id),
		KEY idx_todos_deleted_at (deleted_at),
		KEY todo_deleted_at (deleted_at),
		KEY fk_todos_activity_model (activity_group_id),
		CONSTRAINT fk_todos_activity_model FOREIGN KEY (activity_group_id) REFERENCES activities (id)
	  )
	`

	_, err := DBConn.Exec(activitiesTAble)
	if err != nil {
		return err
	}

	_, err = DBConn.Exec(todosTable)
	if err != nil {
		return err
	}

	return err
}

func InitDatabase(config configs.Conf) error {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.MysqlUser,
		config.MysqlPassword, config.MysqlHost, config.MysqlDBName)
	DBConn, err = sql.Open("mysql", dsn)

	if err != nil {
		return err
	}

	DBConn.SetMaxOpenConns(25)
	DBConn.SetMaxIdleConns(25)
	DBConn.SetConnMaxIdleTime(time.Duration(5 * time.Minute.Minutes()))

	if err := autoMigrate(); err != nil {
		return err
	}

	return nil
}
