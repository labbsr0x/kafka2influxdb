package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labbsr0x/kafka2influxdb/database/models"
	"github.com/labbsr0x/kafka2influxdb/web/config"
	"github.com/labbsr0x/kafka2influxdb/web/services"
	"github.com/labbsr0x/kafka2influxdb/web/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ConsumerController struct {
	*config.WebBuilder
	service *services.ConsumerService
}

func NewConsumerController(webBuilder *config.WebBuilder) *ConsumerController {
	instance := new(ConsumerController)
	instance.service = services.NewConsumerService(webBuilder)
	return instance
}

// CreateHandler saves a single node on influxdb
func (c *ConsumerController) CreateHandler(ctx *gin.Context) {
	var err error
	var json map[string]string
	data := new(models.Data)
	data.Tags = map[string]string{
		"owner": ctx.Param("owner"),
		"thing": ctx.Param("thing"),
		"node":  ctx.Param("node"),
	}

	err = ctx.ShouldBindJSON(&json)
	if err != nil {
		fmt.Printf("Error binding JSON: %s", err)
		ctx.String(http.StatusBadRequest, "Error binding request body JSON. Err:", err)
		return
	}

	if err, data.DateTime = getDateTime(json); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	} else {
		_, servErr := c.service.CreatePoint(data)
		if !servErr.Ok() {
			ctx.String(servErr.SetStatusCode(), fmt.Sprintf("Error saving point: %s", servErr))
			return
		}
	}

	ctx.String(http.StatusCreated, "State point created")
}

// GetHandler retrive a single node from influxdb
func (c *ConsumerController) GetHandler(ctx *gin.Context) {
	var err error
	data := new(models.Data)
	data.Tags = map[string]string{
		"owner": ctx.Param("owner"),
		"thing": ctx.Param("thing"),
		"node":  ctx.Param("node"),
	}

	data.StartDateTime, data.EndDateTime, err = utils.ParsePeriodDateTime(ctx.Query("time"), ctx.Query("startDateTime"), ctx.Query("endDateTime"))
	if err != nil {
		fmt.Printf("%s", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Error parsing time interval query params: %s", err))
		return
	}

	points, servErr := c.service.GetPoints(data)
	if !servErr.Ok() {
		fmt.Printf("%s", servErr)
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Error getting data: %s", servErr))
		return
	}

	println(points)

	ctx.JSON(http.StatusOK, points)
}

func getDateTime(node map[string]string) (error, time.Time) {
	dateTimeString, ok := (node)["dateTime"]
	if !ok {
		logrus.Errorf("dateTime attribute not provied.")
		return fmt.Errorf("The attribute `dateTime` must be provided in the request body. This is the date and time that the data was collected."), time.Time{}
	}

	dateTime, err := time.Parse(time.RFC3339, dateTimeString)
	if err != nil {
		logrus.Errorf("Error parsing dateTime attribute: %s", err)
		return fmt.Errorf("You must provide a `dateTime` field in the RFC3339 format (Ex: 2019-06-24T14:27:33Z)"), time.Time{}
	}
	delete(node, "dateTime")

	return nil, dateTime
}
