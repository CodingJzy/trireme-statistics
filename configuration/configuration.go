package configuration

import (
	"github.com/docopt/docopt-go"
)

const (
	usage = `Collector.
Usage: collector
       [--db-user=<user>]
       [--db-pass=<pass>]
       [--db-name=<name>]
	   [--db-address=<address>]
       [--db-skip-tls=<skiptls>]
       [--ui-user=<user>]
       [--ui-pass=<pass>]
       [--ui-db-address=<address>]
	   [--ui-db-access=<access>]
	   [--listen=<address>]
	   [--log-level=<loglevel>]
       [--log-format=<logformat>]
			 collector -h | --help
			 collector -v | --version
Options:
    -h --help                                   Show this screen.
    -v --version                                Show the version.
    --listen=<address>                          Listening address [default: :8080].

DB Options:
		--db-user=<user>          Username of the database [default: aporeto].
		--db-pass=<pass>         Password of the database [default: aporeto].
		--db-name=<name>         Name of the database [default: flowDB].
		--db-address=<address>   Address to connect to DB [default: http://influxdb:8086].
		--db-skip-tls=<skiptls>  Is valid TLS required for the DB server ? [default: true].

UI Options:
		--ui-user=<user>       Username of the UI to connect with [default: admin].
		--ui-pass=<pass>       Password of the UI to connect with [default: admin].
		--ui-db-address=<address>   Address to connect to UI [default: http://grafana:3000].
		--ui-db-access=<access>   Access to connect to DB [default: proxy].

Log Options:	
		--log-level=<loglevel>    Log level[default: info].
		--log-format=<logformat>    Log format[default: human].
`
)

// Configuration stuct is used to populate the various fields used by collector
type Configuration struct {
	ListenAddress string

	DBUserName string
	DBPassword string
	DBName     string
	DBAddress  string
	DBSkipTLS  bool

	UIUserName string
	UIPassword string
	UIAddress  string
	UIDBAccess string

	LogLevel  string
	LogFormat string
}

// NewConfiguration will parse arguements and return new Configuration struct
func NewConfiguration() *Configuration {

	arguments, err := docopt.Parse(usage, nil, true, "", false)
	if err != nil {
		panic("Unable to parse usage")
	}

	return &Configuration{

		ListenAddress: arguments["--listen"].(string),

		DBUserName: arguments["--db-user"].(string),
		DBPassword: arguments["--db-pass"].(string),
		DBName:     arguments["--db-name"].(string),
		DBAddress:  arguments["--db-address"].(string),
		DBSkipTLS:  arguments["--db-skip-tls"].(bool),
		UIUserName: arguments["--ui-user"].(string),
		UIPassword: arguments["--ui-pass"].(string),
		UIAddress:  arguments["--ui-db-address"].(string),
		UIDBAccess: arguments["--ui-db-access"].(string),
		LogLevel:   arguments["--log-level"].(string),
		LogFormat:  arguments["--log-format"].(string),
	}
}
