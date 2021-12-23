package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/mrrizal/devcode-backend-challenge/configs"
	"github.com/mrrizal/devcode-backend-challenge/database"
	"github.com/mrrizal/devcode-backend-challenge/routes"
)

var Config configs.Conf

func loadEnv() {
	Config.MysqlHost = os.Getenv("MYSQL_HOST")
	Config.MysqlUser = os.Getenv("MYSQL_USER")
	Config.MysqlPassword = os.Getenv("MYSQL_PASSWORD")
	Config.MysqlDBName = os.Getenv("MYSQL_DBNAME")
	Config.Port = 3030
}

func main() {
	loadEnv()
	app := fiber.New()

	if err := database.InitDatabase(Config); err != nil {
		log.Fatal(err.Error())
	}

	routes.SetupRoutes(app)
	app.Listen(fmt.Sprintf(":%d", Config.Port))
}