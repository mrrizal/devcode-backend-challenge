package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/mrrizal/devcode-backend-challenge/cache"
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
	Config.Log = false
	log, err := strconv.ParseBool(os.Getenv("LOG"))
	if err == nil {
		Config.Log = log
	}
	Config.Port = 3030
}

func main() {
	loadEnv()
	app := fiber.New(fiber.Config{
		Prefork:               true,
		DisableStartupMessage: true,
		StreamRequestBody:     true,
		IdleTimeout:           time.Duration(30 * time.Second),
	})

	if Config.Log {
		app.Use(logger.New())
	}

	if err := database.InitDatabase(Config); err != nil {
		log.Fatal(err.Error())
	}

	cache.InitCache()

	routes.SetupRoutes(app)
	app.Listen(fmt.Sprintf(":%d", Config.Port))
}
