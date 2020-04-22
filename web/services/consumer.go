package services

import (
	"fmt"
	"time"

	"github.com/labbsr0x/kafka2influxdb/database/models"
	"github.com/labbsr0x/kafka2influxdb/web/config"
	"github.com/labbsr0x/kafka2influxdb/web/repositories"
	"github.com/labbsr0x/kafka2influxdb/web/utils"
)

type ConsumerService struct {
	repo *repositories.ConsumerRepository
}

func NewConsumerService(webBuilder *config.WebBuilder) *ConsumerService {
	instance := new(ConsumerService)
	instance.repo = repositories.NewConsumerRepository(webBuilder)
	return instance
}

//GetPoint gets an Consumer by primary key
func (s *ConsumerService) GetPoints(data *models.Data) (points []models.StatePoint, servErr utils.ServiceError) {
	var err error
	err = s.ValidateQueryParams(data)

	if err == nil {
		if points, err = s.repo.GetPoints(data); err != nil {
			fmt.Printf("GetPointErr:%s", err)
			if len(points) == 0 {
				servErr.Not_Found = true
			} else {
				servErr.Internal = true
			}
		}
	}
	return
}

// CreatePoint insert a new data
func (s *ConsumerService) CreatePoint(data *models.Data) (body *models.Data, servErr utils.ServiceError) {
	var err error
	err = s.Validate(data)

	if err == nil {
		if err = s.repo.CreatePoint(data); err != nil {
			fmt.Printf("CreatePointErr:%s", err)
			servErr.Internal = true
		} else {
			body = data
		}
	}

	return
}

func (s *ConsumerService) Validate(data *models.Data) error {
	if (data.DateTime == time.Time{}) {
		return fmt.Errorf("The `datetime` parameter is required for sensor data. If you are trying to provide a static data, please prefix the attribute name using the `$` simbol, like: $name, $unity and so on")
	}

	if data.Tags == nil || len(data.Tags) == 0 {
		return fmt.Errorf("The `tags` parameter is required for any data being inserted into AgroWS main database as it is the source of truth and hence must have relevant data")
	}

	if data.Fields == nil || len(data.Fields) == 0 {
		return fmt.Errorf("The `fields` parameter is required for sensor data")
	}

	if data.Tags["owner"] == "" || data.Tags["thing"] == "" || data.Tags["node"] == "" {
		return fmt.Errorf("The `tags` 'owner', 'thing' and 'node' must be provided to every state point being persisted. Tags provided: %s", data.Tags)
	}

	return nil
}

func (s *ConsumerService) ValidateQueryParams(data *models.Data) error {
	if (data.StartDateTime == time.Time{} || data.EndDateTime == time.Time{}) {
		return fmt.Errorf("The `startDateTime` and `endDateTime` parameter are required for querying the database")
	}

	if data.Tags == nil || len(data.Tags) == 0 {
		return fmt.Errorf("The `tags` parameter is required for any data being inserted into AgroWS main database as it is the source of truth and hence must have relevant data")
	}

	if (data.Tags["owner"] == "" && data.Tags["thing"] == "" && data.Tags["node"] == "") || (data.Tags["owner"] == "+" && data.Tags["thing"] == "+" && data.Tags["node"] == "+") {
		return fmt.Errorf("At leats on of the 'owner', 'thing' or 'node' `tags` must be provided for querying the database")
	}

	return nil
}
