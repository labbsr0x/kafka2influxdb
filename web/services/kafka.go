package services

import (
	"encoding/json"
	"fmt"

	"github.com/labbsr0x/kafka2influxdb/web/config"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type KafkaService struct {
	schemaRegistry string
}

func NewKafkaService(webBuilder *config.WebBuilder) *KafkaService {
	instance := new(KafkaService)
	instance.schemaRegistry = webBuilder.KafkaSchemaRegistry
	return instance
}

//GetSchemaID get a schema by id.
func (s *KafkaService) GetSchemaID(topicName string) (int64, error) {
	var schemaID int64
	res, err := doGet([]byte(fmt.Sprintf("%s/subjects/%s/versions/latest/schema", s.schemaRegistry, topicName)), "application/json")
	if err != nil {
		logrus.Errorf("Error at GetSchemaID: %s", err)
		return 0, err
	}
	body := res.Body()
	if res.StatusCode() == fasthttp.StatusOK {
		var response map[string]interface{}
		var found bool
		_ = json.Unmarshal(body, &response)
		id, found := response["id"].(float64)
		if !found {
			logrus.Errorf("ID not found on response GetSchemaID: %s", string(body))
			return 0, fmt.Errorf("ID not found on response GetSchemaID: %s", string(body))
		}
		schemaID = int64(id)
	} else {
		logrus.Errorf("Different response expected!")
		logrus.Errorf("Status Code: %d", res.StatusCode())
		logrus.Errorf("Body: %s", body)
		return 0, fmt.Errorf("Different response expected! Status Code: %d -- Body: %s", res.StatusCode(), string(body))
	}
	return schemaID, nil
}

//LoadSchemaFromRegistry .
func (s *KafkaService) LoadSchemaFromRegistry(schemaID int32) (string, string, error) {
	url := fmt.Sprintf("%s/schemas/ids/%d", s.schemaRegistry, schemaID)
	res, err := doGet([]byte(url), "application/json")
	if err != nil {
		logrus.Errorf("Error at GetSchema ID %d: %s", schemaID, err)
		return "", "", err
	}
	body := res.Body()
	if res.StatusCode() == fasthttp.StatusOK {
		var response map[string]interface{}
		_ = json.Unmarshal([]byte(body), &response)
		logrus.Infof("Response: %s", response)
		schemaModel, found := response["schema"]
		if found {
			err = json.Unmarshal([]byte(schemaModel.(string)), &response)
			if err != nil {
				return fmt.Sprintf("{ \"type\": %s }", schemaModel), "", nil
			}
		}
		return schemaModel.(string), response["name"].(string), nil
	}
	return "", "", fmt.Errorf("Error on aquire Schema ID %d", schemaID)
}

//doGet do a http get request.
func doGet(url []byte, contentType string) (*fasthttp.Response, error) {
	logrus.Infof("URL: %s", string(url))
	req := fasthttp.AcquireRequest()
	req.Header.SetMethodBytes([]byte("GET"))
	req.Header.SetContentType(contentType)
	req.SetRequestURIBytes(url)
	res := fasthttp.AcquireResponse()
	if err := fasthttp.Do(req, res); err != nil {
		logrus.Errorf("Error when make request to kafka rest proxy: %s", err)
		return nil, fmt.Errorf("Error when make request to kafka rest proxy: %s", err)
	}
	fasthttp.ReleaseRequest(req)
	return res, nil
}
