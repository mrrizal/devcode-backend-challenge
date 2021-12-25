package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mrrizal/devcode-backend-challenge/cache"
	"github.com/mrrizal/devcode-backend-challenge/database"
	"github.com/mrrizal/devcode-backend-challenge/models"
	"github.com/mrrizal/devcode-backend-challenge/parser"
	"github.com/mrrizal/devcode-backend-challenge/serializer"
)

func root(c *fiber.Ctx) error {
	resp := serializer.BaseResponse{
		Message: "Welcome to API TODO",
	}
	return c.JSON(resp)
}

func createActivity(c *fiber.Ctx) error {
	db := database.DBConn

	activity := new(models.ActivityModel)
	if err := c.BodyParser(activity); err != nil {
		return parser.GetActivityResponseNoData(c, 400, "Bad Request", "title cannot be null")
	}

	isValid, message := activity.Validate()
	if !isValid {
		return parser.GetActivityResponseNoData(c, 400, "Bad Request", message)
	}

	now := time.Now()
	activity.CreatedAt = now
	activity.UpdatedAt = now

	result := db.Create(&activity)
	if result.Error != nil {
		return parser.GetActivityResponseNoData(c, 500, "Internal Server Error", result.Error.Error())
	}
	return parser.GetActivityResponse(c, 201, "Success", "Success", activity)
}

func getActivity(c *fiber.Ctx) error {
	db := database.DBConn
	activity := new(models.ActivityModel)
	cache := cache.Cache
	expire := 120
	key := []byte(fmt.Sprintf("activity-%s", c.Params("id")))

	got, err := cache.Get(key)
	if err == nil {
		if err := json.Unmarshal(got, &activity); err != nil {
			return parser.GetActivityResponseNoData(c, 500, "Internal Server Error", err.Error())
		}
		return parser.GetActivityResponse(c, 200, "Success", "Success", activity)
	}

	db.Where("deleted_at is null").Find(&activity, c.Params("id"))
	if activity.ID == 0 {
		return parser.GetActivityResponseNoData(c, 404, "Not Found",
			fmt.Sprintf("Activity with ID %s Not Found", c.Params("id")))
	}

	// set cache
	activitiesBytes := new(bytes.Buffer)
	json.NewEncoder(activitiesBytes).Encode(activity)
	if err := cache.Set(key, activitiesBytes.Bytes(), expire); err != nil {
		return parser.GetActivityResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	return parser.GetActivityResponse(c, 200, "Success", "Success", activity)
}

func getActivities(c *fiber.Ctx) error {
	db := database.DBConn
	cache := cache.Cache
	var activities []*models.ActivityModel
	expire := 120

	key := []byte("activities")
	var firstID, lastID struct {
		ID int
	}

	// get data from cache
	got, err := cache.Get(key)
	if err == nil {
		if err := json.Unmarshal(got, &activities); err != nil {
			return parser.GetActivityResponseNoData(c, 500, "Internal Server Error", err.Error())
		}
		return parser.GetActivitiesResponse(c, 200, "Success", "Success", activities)
	}

	bucketSize := 1000
	resultCount := 0
	db.Model(&models.ActivityModel{}).First(&firstID)
	db.Model(&models.ActivityModel{}).Last(&lastID)

	resultChannel := make(chan []*models.ActivityModel)
	for beginID := firstID.ID; beginID <= lastID.ID; beginID += bucketSize {
		endID := beginID + bucketSize
		go func(beginID, endID int) {
			var tempActivities []*models.ActivityModel
			db.Where("deleted_at is null and id >= ? and id < ?", beginID, endID).Find(&tempActivities)
			resultChannel <- tempActivities
		}(beginID, endID)
		resultCount += 1
	}

	for i := 0; i < resultCount; i++ {
		tempActivities := <-resultChannel
		activities = append(activities, tempActivities...)
	}
	sort.Slice(activities, func(i, j int) bool { return activities[i].ID < activities[j].ID })

	// set cache
	activitiesBytes := new(bytes.Buffer)
	json.NewEncoder(activitiesBytes).Encode(activities)
	if err := cache.Set(key, activitiesBytes.Bytes(), expire); err != nil {
		return parser.GetActivityResponseNoData(c, 500, "Internal Server Error", err.Error())
	}

	return parser.GetActivitiesResponse(c, 200, "Success", "Success", activities)
}

func updateActivity(c *fiber.Ctx) error {
	db := database.DBConn
	title := struct {
		Title string `json:"title"`
	}{}

	if err := c.BodyParser(&title); err != nil {
		return parser.GetActivityResponseNoData(c, 400, "Bad Request", "title cannot be null")
	}

	var activity *models.ActivityModel
	db.Where("deleted_at is null").Find(&activity, c.Params("id"))
	if activity.ID == 0 {
		return parser.GetActivityResponseNoData(c, 404, "Not Found",
			fmt.Sprintf("Activity with ID %s Not Found", c.Params("id")))
	}
	activity.Title = title.Title
	activity.UpdatedAt = time.Now().UTC()
	db.Save(&activity)

	cache := cache.Cache
	cache.Del([]byte(fmt.Sprintf("activity-%s", c.Params("id"))))
	return parser.GetActivityResponse(c, 200, "Success", "Success", activity)
}

func deleteActivity(c *fiber.Ctx) error {
	db := database.DBConn
	resp := db.Model(&models.ActivityModel{}).Where("id = ? and deleted_at is null", c.Params("id")).
		Update("deleted_at", time.Now().UTC())

	if resp.Error != nil || resp.RowsAffected == 0 {
		return parser.GetActivityResponseNoData(c, 404, "Not Found", fmt.Sprintf("Activity with ID %s Not Found",
			c.Params("id")))
	}

	return parser.GetActivityResponseNoData(c, 200, "Success", "Success")
}

func SetupRoutes(app *fiber.App) {
	app.Get("/", root)
	app.Post("/activity-groups/", createActivity)
	app.Get("/activity-groups/", getActivities)
	app.Get("/activity-groups/:id", getActivity)
	app.Patch("/activity-groups/:id", updateActivity)
	app.Delete("/activity-groups/:id", deleteActivity)
}
