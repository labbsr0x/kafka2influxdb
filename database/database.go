package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/labbsr0x/kafka2influxdb/database/models"
	"github.com/labbsr0x/kafka2influxdb/web/config"

	"github.com/huandu/go-sqlbuilder"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/sirupsen/logrus"
)

// Database defines the methods that can be performed
type Database interface {
	Init(webBuilder *config.WebBuilder) Database
	Connect() Database
	Close() error
	GetPoints(data *models.Data) ([]models.StatePoint, error)
	CreatePoint(data *models.Data) (*client.Point, error)
}

// DefaultDatabase a default Database interface implementation
type DefaultDatabase struct {
	Name      string
	Addr      string
	User      string
	Password  string
	Precision string
	Client    client.Client
}

// Init initializes a default credentials DAO from web builder
func (db *DefaultDatabase) Init(webBuilder *config.WebBuilder) Database {
	db.Name = webBuilder.InfluxdbName
	db.Addr = webBuilder.InfluxdbAddr
	db.User = webBuilder.InfluxdbUser
	db.Password = webBuilder.InfluxdbPassword
	db.Precision = "s"

	return db
}

// Connect to influxDB
func (db *DefaultDatabase) Connect() Database {
	logrus.Debugf("Connecting to InfluxDB. Host: %s, Username: %s, Password: %s", db.Addr, db.User, db.Password)

	httpClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     db.Addr,
		Username: db.User,
		Password: db.Password,
	})

	if err != nil {
		logrus.Errorf("error connecting to influxDB: %s", err)
		panic(fmt.Sprintf("No error should happen when connecting to database, but got err=%+v", err))
	}

	db.Client = httpClient
	logrus.Debugf("connected to influxDB")

	return db
}

// Close all opened connections
func (db *DefaultDatabase) Close() error {
	return db.Client.Close()
}

// Saves a point to Influx. A point represents a state of a sensor in time
func (db *DefaultDatabase) CreatePoint(data *models.Data) (*client.Point, error) {
	attributes := map[string]interface{}{}

	influxPoints, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  db.Name,
		Precision: db.Precision,
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating new batch point. Details: %s ", err)
	}

	for k, v := range data.Fields {
		// check whether there is a static data attribute
		if strings.Contains(k, "$") {
			return nil, fmt.Errorf(fmt.Sprintf("Static attributes (aka the ones prefixed with `$)` must be saved in the static data DataBase as it doesn't change over time. Invalid attributed: %s", k))
		}
		attributes[k] = v
	}

	influxPoint, err := client.NewPoint("state", data.Tags, attributes, data.DateTime)
	if err != nil {
		return nil, fmt.Errorf("Error creating new point. Details %s ", err)
	}
	influxPoints.AddPoint(influxPoint)

	err = db.Client.Write(influxPoints)

	if err != nil {
		return nil, fmt.Errorf("Error writing the batch points to influx. Details: %s", err)
	}

	return influxPoint, nil
}

// Retrieves a point in time
func (db *DefaultDatabase) GetPoints(data *models.Data) ([]models.StatePoint, error) {
	var values [][]interface{}
	var ownerIndex int
	var thingIndex int
	var nodeIndex int

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("*")
	sb.From("state")

	if (data.StartDateTime != time.Time{}) {
		sb.Where(sb.And(sb.GreaterEqualThan("time", data.StartDateTime.Format(time.RFC3339))))
	}
	if (data.EndDateTime != time.Time{}) {
		sb.Where(sb.And(sb.LessEqualThan("time", data.EndDateTime.Format(time.RFC3339))))
	}
	if data.Tags["owner"] != "" && data.Tags["owner"] != "+" {
		sb.Where(sb.And(sb.Equal("owner", data.Tags["owner"])))
	}
	if data.Tags["thing"] != "" && data.Tags["thing"] != "+" {
		sb.Where(sb.And(sb.Equal("thing", data.Tags["thing"])))
	}
	if data.Tags["node"] != "" && data.Tags["node"] != "+" {
		sb.Where(sb.And(sb.Equal("node", data.Tags["node"])))
	}

	sql, args := sb.Build()
	queryFmt := fmt.Sprintf(strings.ReplaceAll(sql, "?", "'%s'"), args...)

	q := client.Query{
		Command:  queryFmt,
		Database: db.Name,
	}

	response, err := db.Client.Query(q)
	if err != nil {
		return nil, fmt.Errorf("Error querying state points. Details: %s", err)
	}

	if response.Error() != nil {
		return nil, fmt.Errorf("Error quering Influx for state points. Details: %s", response.Error())
	}

	if len(response.Results) == 0 || len(response.Results[0].Series) == 0 {
		return make([]models.StatePoint, 0), nil
	}

	values = response.Results[0].Series[0].Values
	cols := response.Results[0].Series[0].Columns

	attributeMapping := map[string]int{}
	for j, c := range cols {
		if c == "owner" {
			ownerIndex = j
		} else if c == "thing" {
			thingIndex = j
		} else if c == "node" {
			nodeIndex = j
		} else if c != "time" {
			attributeMapping[c] = j
		}
	}

	r := make([]models.StatePoint, len(values))
	for i, v := range values {
		dt, err := time.Parse(time.RFC3339, v[0].(string))
		if err != nil {
			return nil, fmt.Errorf("Error parsing dateTime for result. Error details: %s", err)
		}
		attributes := map[string]string{}
		for key, val := range attributeMapping {
			if v[val] != nil {
				attributes[key] = v[val].(string)
			}
		}
		r[i] = models.StatePoint{DateTime: dt, Owner: v[ownerIndex].(string), Thing: v[thingIndex].(string), Node: v[nodeIndex].(string), Attributes: attributes}
	}

	return r, nil
}
