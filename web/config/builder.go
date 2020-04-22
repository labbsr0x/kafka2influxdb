package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	kafkaAddr        = "kafka-addr"
	kafkaTopic       = "kafka-topic"
	kafkaPartition   = "kafka-partition"
	influxdbAddr     = "influxdb-addr"
	influxdbName     = "influxdb-name"
	influxdbUser     = "influxdb-user"
	influxdbPassword = "influxdb-password"
	port             = "port"
	logLevel         = "log-level"
)

// Flags define the fields that will be passed via cmd
type Flags struct {
	KafkaAddr        string
	KafkaTopic       string
	KafkaPartition   int
	InfluxdbName     string
	InfluxdbAddr     string
	InfluxdbUser     string
	InfluxdbPassword string
	Port             string
	LogLevel         string
}

// WebBuilder defines the parametric information of a server instance
type WebBuilder struct {
	*Flags
}

// AddFlags adds flags for Builder.
func AddFlags(flags *pflag.FlagSet) {
	flags.StringP(kafkaAddr, "k", "", "Kafka URL")
	flags.StringP(kafkaTopic, "t", "/owner/*", "Kafka's topic")
	flags.StringP(kafkaPartition, "", "0", "[optional] Kafka's partition. Default: 0")
	flags.StringP(influxdbAddr, "i", "", "InfluxDB URL")
	flags.StringP(influxdbName, "n", "interactws", "[optional] Sets the InfluxDB's name. Default: 'interactws'")
	flags.StringP(influxdbUser, "u", "", "Sets the InfluxDB's user")
	flags.StringP(influxdbPassword, "s", "", "Sets the InfluxDB's password")
	flags.StringP(port, "p", "7070", "[optional] Custom port for accessing Kafka2InfluxDB's services. Defaults to 7070")
	flags.StringP(logLevel, "l", "info", "[optional] Sets the Log Level to one of seven (trace, debug, info, warn, error, fatal, panic). Defaults to info")
}

// Init initializes the web server builder with properties retrieved from Viper.
func (b *WebBuilder) Init(v *viper.Viper) *WebBuilder {
	flags := new(Flags)
	flags.KafkaAddr = v.GetString(kafkaAddr)
	flags.KafkaTopic = v.GetString(kafkaTopic)
	flags.KafkaPartition = v.GetInt(kafkaPartition)
	flags.InfluxdbAddr = v.GetString(influxdbAddr)
	flags.InfluxdbName = v.GetString(influxdbName)
	flags.InfluxdbUser = v.GetString(influxdbUser)
	flags.InfluxdbPassword = v.GetString(influxdbPassword)
	flags.Port = v.GetString(port)
	flags.LogLevel = v.GetString(logLevel)
	flags.check()

	b.Flags = flags

	return b
}

func (flags *Flags) check() {
	logrus.Infof("Flags: '%v'", flags)

	requiredFlags := []struct {
		value string
		name  string
	}{
		{flags.KafkaAddr, kafkaAddr},
		{flags.KafkaTopic, kafkaTopic},
		{flags.InfluxdbAddr, influxdbAddr},
		{flags.InfluxdbUser, influxdbUser},
		{flags.InfluxdbPassword, influxdbPassword},
	}

	var errMsg string

	for _, flag := range requiredFlags {
		if flag.value == "" {
			errMsg += fmt.Sprintf("\n\t%v", flag.name)
		}
	}

	if errMsg != "" {
		errMsg = "The following flags are missing: " + errMsg
		panic(errMsg)
	}
}
