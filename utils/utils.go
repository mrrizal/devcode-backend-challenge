package utils

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/mrrizal/devcode-backend-challenge/models"
)

func GetActivityErrorResponse(c *fiber.Ctx, statusCode int, status, message string) error {
	resp := models.ActivityErrorResponse{
		Status:  status,
		Message: message,
		Data:    make(map[string]string),
	}
	return json.NewEncoder(c.Type("json", "utf-8").Status(statusCode).Response().BodyWriter()).Encode(resp)

}

func GetActivityResponse(c *fiber.Ctx, statusCode int, status, message string, activity *models.Activity) error {
	resp := models.ActivityResponse{
		Status:  status,
		Message: message,
		Data:    activity,
	}
	return json.NewEncoder(c.Type("json", "utf-8").Status(statusCode).Response().BodyWriter()).Encode(resp)
}
