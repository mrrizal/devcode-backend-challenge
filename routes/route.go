package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mrrizal/devcode-backend-challenge/models"
	"github.com/mrrizal/devcode-backend-challenge/utils"
)

func root(c *fiber.Ctx) error {
	resp := models.BaseResponse{
		Message: "Welcome to API TODO",
	}
	return c.JSON(resp)
}

func createActivity(c *fiber.Ctx) error {
	activity := new(models.Activity)
	if err := c.BodyParser(activity); err != nil {
		return utils.GetActivityErrorResponse(c, 400, "Bad Request", "title cannot be null")

	}

	if !activity.Validate() {
		return utils.GetActivityErrorResponse(c, 400, "Bad Request", "title cannot be null")
	}

	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()
	return utils.GetActivityResponse(c, 201, "Success", "Success", activity)
}

func SetupRoutes(app *fiber.App) {
	app.Get("/", root)
	app.Post("/activity-groups/", createActivity)
}
