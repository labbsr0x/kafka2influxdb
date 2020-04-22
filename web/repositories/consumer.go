package repositories

import (
	"fmt"

	"github.com/labbsr0x/kafka2influxdb/database"
	"github.com/labbsr0x/kafka2influxdb/database/models"
	"github.com/labbsr0x/kafka2influxdb/web/config"
)

type ConsumerRepository struct {
	db database.Database
}

func NewConsumerRepository(webBuilder *config.WebBuilder) *ConsumerRepository {
	instance := new(ConsumerRepository)
	instance.db = new(database.DefaultDatabase).Init(webBuilder)
	return instance
}

func (r *ConsumerRepository) GetPoints(element *models.Data) (points []models.StatePoint, err error) {
	if points, err = r.db.Connect().GetPoints(element); err != nil {
		fmt.Printf("GetPointErr:%s", err)
	}

	r.db.Close()
	return
}

func (r *ConsumerRepository) CreatePoint(element *models.Data) (err error) {
	if _, err = r.db.Connect().CreatePoint(element); err != nil {
		fmt.Printf("CreatePointErr:%s", err)
	}

	r.db.Close()
	return
}
