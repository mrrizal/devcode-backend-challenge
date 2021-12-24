package parser

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/mrrizal/devcode-backend-challenge/models"
	"github.com/mrrizal/devcode-backend-challenge/serializer"
)

func parseActivity(activityModel *models.ActivityModel) serializer.Activity {
	activity := serializer.Activity{
		ID:        activityModel.ID,
		Email:     activityModel.Email,
		Title:     activityModel.Title,
		CreatedAt: fmt.Sprintf("%sZ", activityModel.CreatedAt.Format("2006-01-02T15:04:05.000")),
		UpdatedAt: fmt.Sprintf("%sZ", activityModel.UpdatedAt.Format("2006-01-02T15:04:05.000")),
	}
	if activityModel.DeletedAt != nil {
		deletedAt := *activityModel.DeletedAt
		*activity.DeletedAt = fmt.Sprintf("%sZ", deletedAt.Format("2006-01-02T15:04:05.000"))
	}
	return activity
}

func GetActivityErrorResponse(c *fiber.Ctx, statusCode int, status, message string) error {
	resp := serializer.ActivityErrorResponse{
		Status:  status,
		Message: message,
		Data:    make(map[string]string),
	}
	return json.NewEncoder(c.Type("json", "utf-8").Status(statusCode).Response().BodyWriter()).Encode(resp)

}

func GetActivityResponse(c *fiber.Ctx, statusCode int, status, message string, activityModel *models.ActivityModel) error {
	activity := parseActivity(activityModel)
	resp := serializer.ActivityResponse{
		Status:  status,
		Message: message,
		Data:    &activity,
	}
	return json.NewEncoder(c.Type("json", "utf-8").Status(statusCode).Response().BodyWriter()).Encode(resp)
}

func GetActivitiesResponse(c *fiber.Ctx, statusCode int, status, message string, activitiesModel []*models.ActivityModel) error {
	var activities []serializer.Activity
	for _, activityModel := range activitiesModel {
		activity := parseActivity(activityModel)
		activities = append(activities, activity)
	}

	resp := serializer.ActivitiesResponse{
		Status:  status,
		Message: message,
		Data:    &activities,
	}
	return json.NewEncoder(c.Type("json", "utf-8").Status(statusCode).Response().BodyWriter()).Encode(resp)
}
