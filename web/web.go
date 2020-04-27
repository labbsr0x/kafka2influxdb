package web

import (
	"net/http"

	"github.com/labbsr0x/kafka2influxdb/database"
	"github.com/labbsr0x/kafka2influxdb/web/config"
	"github.com/labbsr0x/kafka2influxdb/web/controllers"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Server struct {
	*config.WebBuilder
	app      *gin.Engine
	consumer *controllers.ConsumerController
	kafka    *database.DefaultKafka
}

// InitFromWebBuilder builds a Server instance
func (s *Server) InitFromWebBuilder(webBuilder *config.WebBuilder) *Server {
	s.WebBuilder = webBuilder
	s.app = gin.Default()
	s.consumer = controllers.NewConsumerController(s.WebBuilder)
	s.kafka = database.NewKafka(s.WebBuilder).Connect()

	logLevel, err := logrus.ParseLevel(s.WebBuilder.LogLevel)
	if err != nil {
		logrus.Errorf("Not able to parse log level string. Setting default level: info.")
		logLevel = logrus.InfoLevel
	}
	logrus.SetLevel(logLevel)

	return s
}

func (s *Server) Run() {
	consumerGroup := s.app.Group("/")
	{
		consumerGroup.GET("/", index)
		consumerGroup.GET("/owner/:owner/thing/:thing/node/:node", s.consumer.GetHandler)
		consumerGroup.POST("/owner/:owner/thing/:thing/node/:node", s.consumer.CreateHandler)
	}

<<<<<<< HEAD
	go s.kafka.ListenGroup(s.consumer.ListenHandler)

=======
	s.kafka.ListenGroup(s.consumer.ListenHandler)
>>>>>>> Decoding avro
	s.app.Run("0.0.0.0:" + s.WebBuilder.Port)
}

func index(ctx *gin.Context) {
	ctx.String(http.StatusOK, "Welcome to Kafka to InfluxDB Service")
}
