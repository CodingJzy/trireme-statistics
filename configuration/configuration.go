package configuration

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	flag "github.com/spf13/pflag"
)

// Configuration stuct is used to populate the various fields used by trireme-statistics
type Configuration struct {
	ListenAddress string

	InfluxUsername string
	InfluxPassword string
	InfluxDBName   string
	InfluxURL      string
	DBSkipTLS      bool

	GrafanaUsername string
	GrafanaPassword string
	GrafanaURL      string
	GrafanaDBAccess string

	GraphGenerationInterval int

	LogFormat string
	LogLevel  string
}

func usage() {
	flag.PrintDefaults()
	os.Exit(2)
}

// LoadConfiguration will load the configuration struct
func LoadConfiguration() (*Configuration, error) {
	flag.Usage = usage
	flag.String("ListenAddress", "", "Server Address [Default: 8080]")
	flag.String("LogLevel", "", "Log level. Default to info (trace//debug//info//warn//error//fatal)")
	flag.String("LogFormat", "", "Log Format. Default to human")

	flag.String("InfluxUsername", "", "Username of the database [default: aporeto]")
	flag.String("InfluxPassword", "", "Password of the database [default: aporeto]")
	flag.String("InfluxDBName", "", "Name of the database [default: flowDB]")
	flag.String("InfluxURL", "", "URI to connect to DB [default: http://influxdb:8086]")
	flag.Bool("DBSkipTLS", true, "Is valid TLS required for the DB server ? [default: true]")

	flag.String("GrafanaUsername", "", "Username of the UI to connect with [default: admin]")
	flag.String("GrafanaPassword", "", "Password of the UI to connect with [default: admin]")
	flag.String("GrafanaURL", "", "URI to connect to UI [default: http://grafana:3000]")
	flag.String("GrafanaDBAccess", "", "Access to connect to DB [default: proxy]")

	flag.Int("GraphGenerationInterval", 20, "Time interval between transformation [default: 20s]")

	// Setting up default configuration
	viper.SetDefault("ListenAddress", ":8080")
	viper.SetDefault("LogLevel", "info")
	viper.SetDefault("LogFormat", "human")

	viper.SetDefault("InfluxUsername", "aporeto")
	viper.SetDefault("InfluxPassword", "aporeto")
	viper.SetDefault("InfluxDBName", "flowDB")
	viper.SetDefault("InfluxURL", "http://influxdb:8086")
	viper.SetDefault("DBSkipTLS", true)

	viper.SetDefault("GrafanaUsername", "admin")
	viper.SetDefault("GrafanaPassword", "admin")
	viper.SetDefault("GrafanaURL", "http://grafana:3000")
	viper.SetDefault("GrafanaDBAccess", "proxy")

	viper.SetDefault("GraphGenerationInterval", 20)

	// Binding ENV variables
	// Each config will be of format TRIREME_XYZ as env variable, where XYZ
	// is the upper case config.
	viper.SetEnvPrefix("TRIREME")
	viper.AutomaticEnv()

	// Binding CLI flags.
	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	var config Configuration

	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling: %s", err)
	}

	return &config, nil
}
