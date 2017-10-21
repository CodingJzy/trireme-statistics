package configuration

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	flag "github.com/spf13/pflag"
)

// DefaultKubeConfigLocation is the default location of the KubeConfig file.
const DefaultKubeConfigLocation = "/.kube/config"

// Configuration stuct is used to populate the various fields used by collector
type Configuration struct {
	ListenAddress  string
	KubeconfigPath string

	DBUserName string
	DBPassword string
	DBName     string
	DBAddress  string
	DBPort     string
	DBIP       string
	DBSkipTLS  bool

	UIUserName string
	UIPassword string
	UIAddress  string
	UIDBAccess string
	UIPort     string
	UIIP       string

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
	flag.String("KubeconfigPath", "", "KubeConfig used to connect to Kubernetes")
	flag.String("LogLevel", "", "Log level. Default to info (trace//debug//info//warn//error//fatal)")
	flag.String("LogFormat", "", "Log Format. Default to human")

	flag.String("DBUserName", "", "Username of the database [default: aporeto]")
	flag.String("DBPassword", "", "Password of the database [default: aporeto]")
	flag.String("DBName", "", "Name of the database [default: flowDB]")
	flag.String("DBIP", "", "IP address of the database [default: influxdb]")
	flag.String("DBPort", "", "Port of the database [default: 8086]")
	flag.String("DBAddress", "", "URI to connect to DB [default: http://influxdb:8086]")
	flag.String("DBSkipTLS", "", "Is valid TLS required for the DB server ? [default: true]")

	flag.String("UIUserName", "", "Username of the UI to connect with [default: admin]")
	flag.String("UIPassword", "", "Password of the UI to connect with [default: admin]")
	flag.String("UIIP", "", "IP address of the UI [default: grafana]")
	flag.String("UIPort", "", "Port of the UI [default: 3000]")
	flag.String("UIAddress", "", "URI to connect to UI [default: http://grafana:3000]")
	flag.String("UIDBAccess", "", "Access to connect to DB [default: proxy]")

	// Setting up default configuration
	viper.SetDefault("ListenAddress", ":8080")
	viper.SetDefault("KubeconfigPath", "")
	viper.SetDefault("LogLevel", "info")
	viper.SetDefault("LogFormat", "human")

	viper.SetDefault("DBUserName", "aporeto")
	viper.SetDefault("DBPassword", "aporeto")
	viper.SetDefault("DBName", "flowDB")
	viper.SetDefault("DBIP", "influxdb")
	viper.SetDefault("DBPort", ":8086")
	viper.SetDefault("DBAddress", "http://influxdb:8086")
	viper.SetDefault("DBSkipTLS", true)

	viper.SetDefault("UIUserName", "admin")
	viper.SetDefault("UIPassword", "admin")
	viper.SetDefault("UIIP", "grafana")
	viper.SetDefault("UIPort", ":3000")
	viper.SetDefault("UIAddress", "http://grafana:3000")
	viper.SetDefault("UIDBAccess", "proxy")

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
		return nil, fmt.Errorf("error unmarshalling:%s", err)
	}

	err = validateConfig(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// validateConfig is validating the Configuration struct.
func validateConfig(config *Configuration) error {
	// Validating KUBECONFIG
	// In case not running as InCluster, we try to infer a possible KubeConfig location
	if os.Getenv("KUBERNETES_PORT") == "" {
		if config.KubeconfigPath == "" {
			config.KubeconfigPath = os.Getenv("HOME") + DefaultKubeConfigLocation
		}
	} else {
		config.KubeconfigPath = ""
	}

	return nil
}
