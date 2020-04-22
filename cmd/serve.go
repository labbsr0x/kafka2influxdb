package cmd

import (
	"github.com/labbsr0x/kafka2influxdb/web"
	"github.com/labbsr0x/kafka2influxdb/web/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the HTTP REST APIs server",
	RunE: func(cmd *cobra.Command, args []string) error {
		builder := new(config.WebBuilder).Init(viper.GetViper())
		server := new(web.Server).InitFromWebBuilder(builder)
		server.Run()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	config.AddFlags(serveCmd.Flags())

	err := viper.GetViper().BindPFlags(serveCmd.Flags())
	if err != nil {
		panic(err)
	}
}
