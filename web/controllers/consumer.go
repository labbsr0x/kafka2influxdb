package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/labbsr0x/kafka2influxdb/database/models"
	"github.com/labbsr0x/kafka2influxdb/web/config"
	"github.com/labbsr0x/kafka2influxdb/web/services"
	"github.com/labbsr0x/kafka2influxdb/web/utils"

	"github.com/gin-gonic/gin"
	"github.com/hamba/avro"
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

// ListenHandler saves a single node on influxdb
func (c *ConsumerController) ListenHandler(payload []byte) error {
	data, err := getData(payload)
	if err != nil {
		logrus.Errorf("Error binding JSON: %s", err)
		return fmt.Errorf("Error binding JSON: %s", err)
	}

	_, servErr := c.service.CreatePoint(data)
	if !servErr.Ok() {
		logrus.Errorf("Error saving point: %s", servErr)
		return fmt.Errorf("Error saving point: %s", servErr)
	}

	return nil
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
		logrus.Errorf("Error binding JSON: %s", err)
		ctx.String(http.StatusBadRequest, "Error binding request body JSON. Err:", err)
		return
	}

	if err, data.DateTime = getDateTime(json); err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	} else {
		data.Fields = json
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
		logrus.Errorf("%s", err)
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Error parsing time interval query params: %s", err))
		return
	}

	points, servErr := c.service.GetPoints(data)
	if !servErr.Ok() {
		logrus.Errorf("%s", servErr)
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
		return fmt.Errorf("You must provide a `dateTime` field in the RFC3339 format (Ex: 2020-05-24T14:27:33Z)"), time.Time{}
	}
	delete(node, "dateTime")

	return nil, dateTime
}

func getData(payload []byte) (data *models.Data, err error) {
	var schema avro.Schema
	var dateTime time.Time
	record := models.Schema{}
	retry := true

	schema, err = avro.Parse(models.SchemaModel)
	if err != nil {
		logrus.Errorf("The schema could not be parsed: %v", err)
	}

	err = avro.Unmarshal(schema, payload, &record)
	if err != nil {
		err = json.Unmarshal(payload, &record) //Try decoding as json when avro fails
		if err != nil {
			logrus.Errorf("The message could not be decoded: %v", err)
			return
		}
	}

	logrus.Debugf("Record parsed: %v", record)

	for retry {
		dateTime, err = time.Parse(time.RFC3339, record.Value.DateTime)
		if err != nil {
			err = json.Unmarshal(payload, &record) //Try decoding as json when couldn`t parse dateTime
			if err != nil {
				retry = false
				logrus.Errorf("Error parsing dateTime attribute: %s", err)
				return
			}
		} else {
			retry = false
		}
	}

	data = new(models.Data)
	data.DateTime = dateTime

	rg := regexp.MustCompile(`owner/(?P<Owner>\w+)/thing/(?P<Thing>\w+)/node/(?P<Node>\w+)`)
	if !rg.MatchString(record.Key) {
		err = fmt.Errorf("The keys doesn't matches with pattern (owner/:owner/thing/:thing/node/:node): %s", record.Key)
		logrus.Errorf("Error on parsing tags: %s", err)
		return
	}
	keys := rg.FindStringSubmatch(record.Key)

	data.Tags = map[string]string{
		"owner": keys[1],
		"thing": keys[2],
		"node":  keys[3],
	}

	data.Fields = map[string]string{
		"lat":  record.Value.Lat,
		"lon":  record.Value.Lon,
		"mci":  record.Value.Mci,
		"type": record.Value.Type,
	}

	return
}
