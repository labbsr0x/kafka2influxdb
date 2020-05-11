package controllers

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
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
	service      *services.ConsumerService
	kafkaService *services.KafkaService
}

func NewConsumerController(webBuilder *config.WebBuilder) *ConsumerController {
	instance := new(ConsumerController)
	instance.service = services.NewConsumerService(webBuilder)
	instance.kafkaService = services.NewKafkaService(webBuilder)
	return instance
}

// ListenHandler saves a single node on influxdb
func (c *ConsumerController) ListenHandler(key []byte, payload []byte) error {
	data, err := c.getData(key, payload)
	if err != nil {
		logrus.Errorf("Error binding JSON: %s", err)
		return fmt.Errorf("Error binding JSON: %s", err)
	}

	_, servErr := c.service.CreatePoint(data)
	if !servErr.Ok() {
		logrus.Errorf("Error saving point: %v", servErr)
		return fmt.Errorf("Error saving point: %v", servErr)
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
			ctx.String(servErr.SetStatusCode(), fmt.Sprintf("Error saving point: %v", servErr))
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
		logrus.Errorf("%v", servErr)
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Error getting data: %v", servErr))
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

// https://docs.confluent.io/current/schema-registry/serdes-develop/index.html#wire-format
func (c *ConsumerController) getSchemaOfMessage(payload []byte) (string, string, error) {
	var schemaID int32
	logrus.Tracef("Payload Schema ID: %s", hex.EncodeToString(payload[1:5]))
	buf := bytes.NewBuffer(payload[1:5])
	binary.Read(buf, binary.BigEndian, &schemaID)
	logrus.Debugf("SchemaID: %d\n", schemaID)
	return c.kafkaService.LoadSchemaFromRegistry(schemaID)
}

func (c *ConsumerController) decodeAvro(payload []byte, result interface{}) (string, error) {
	schema, schemaName, err := c.getSchemaOfMessage(payload)
	if err != nil {
		logrus.Errorf("Cannot parse schemaKey: %s", err)
		return "", fmt.Errorf("Cannot parse schemaKey: %s", err)
	}
	logrus.Tracef(schema)

	//When using SchemaRegistry, the message comes with this pattern
	// 0     : Magic byte
	// 1-4   : 4 bytes Schema ID using BigEndian
	// 5-... : Serialized data for the specified schema format
	//https://docs.confluent.io/current/schema-registry/serdes-develop/index.html#wire-format
	msgPayload := payload[5:]
	logrus.Debugf("msgPayload: %s", hex.EncodeToString(msgPayload))

	schemaModel, err := avro.Parse(schema)
	if err != nil {
		logrus.Errorf("The schema could not be parsed: %v", err)
		return schemaName, fmt.Errorf("The schema could not be parsed: %v", err)
	}
	err = avro.Unmarshal(schemaModel, msgPayload, result)
	if err != nil {
		//Try decoding as json when avro fails
		err = json.Unmarshal(payload, result)
		if err != nil {
			logrus.Errorf("The message could not be decoded: %v", err)
			return schemaName, fmt.Errorf("The message could not be decoded: %v", err)
		}
	}
	return schemaName, nil
}

func (c *ConsumerController) getData(keyPayload []byte, messagePayload []byte) (data *models.Data, err error) {
	var dateTime time.Time

	//Parse Key Payload
	var messageKey string
	_, err = c.decodeAvro(keyPayload, &messageKey)
	if err != nil {
		logrus.Errorf("Error decoding key payload: %s", err)
		return
	}
	logrus.Debugf("Key parsed: %s", messageKey)

	//Parse Message Payload
	var message map[string]interface{}
	messageSchemaName, err := c.decodeAvro(messagePayload, &message)
	if err != nil {
		logrus.Errorf("Error decoding message payload: %s", err)
		return
	}
	logrus.Debugf("Record parsed: %s", message)

	dateTime, err = time.Parse(time.RFC3339, message["dateTime"].(string))
	if err != nil {
		dateTime, err = time.Parse("2006-01-02T15:04:05Z0700", message["dateTime"].(string))
		if err != nil {
			logrus.Errorf("Error on parse dateTime: %s", err)
			return
		}
	}

	data = new(models.Data)
	data.DateTime = dateTime

	rg := regexp.MustCompile(`owner/(?P<Owner>\w+)/thing/(?P<Thing>\w+)/node/(?P<Node>\w+)`)
	if !rg.MatchString(messageKey) {
		err = fmt.Errorf("The keys doesn't matches with pattern (owner/:owner/thing/:thing/node/:node): %s", messageKey)
		logrus.Errorf("Error on parsing tags: %s", err)
		return
	}
	keys := rg.FindStringSubmatch(messageKey)

	data.Tags = map[string]string{
		"owner":    keys[1],
		"thing":    keys[2],
		"node":     keys[3],
		"schema_0": messageSchemaName,
	}

	data.Fields = map[string]string{
		"lat":  message["lat"].(string),
		"lon":  message["lon"].(string),
		"mci":  message["mci"].(string),
		"type": message["type"].(string),
	}

	return
}

func (c *ConsumerController) getSchemaId(topic string) (schemaID int64, err error) {
	schemaID, err = c.kafkaService.GetSchemaID(topic)
	if err != nil {
		logrus.Errorf("Error recovering schema id: %s", err)
		return
	}

	return
}
