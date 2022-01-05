package configs

type Conf struct {
	MysqlHost     string
	MysqlUser     string
	MysqlPassword string
	MysqlDBName   string
	Port          int
	Log           bool
}
