package utils

import (
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	RedisHost            string             `yaml:"redisHost" json:"redisHost"`
	RedisRFSHost         *RedisFailOverConf `yaml:"redisRFSHost" json:"redisRFSHost"`
	GatewayHost          string             `yaml:"gatewayHost" json:"gatewayHost"`
	Namespace            string             `yaml:"namespace" json:"namespace"`
	ConcurrencyLimit     int64              `yaml:"concurrency_limit" json:"concurrency_limit"`
	GatewayJWTToken      string             `yaml:"gateway_jwt_token" json:"gateway_jwt_token"`
	MatrixTimeoutSeconds int64              `yaml:"matrix_timeout_seconds" json:"matrix_timeout_seconds"`
	OpenAPIDocPath       string             `yaml:"openapi_doc_path" json:"openapi_doc_path"`
	TokenAuds            map[string]bool    `yaml:"token_auds" json:"token_auds"`
	JobIDPrefix          string             `yaml:"job_id_prefix" json:"job_id_prefix"`
	Cluster              string             `yaml:"cluster" json:"cluster"`
	PubsubTopic          string             `yaml:"pubsub_topic" json:"pubsub_topic"`
	CacheId              bool               `yaml:"cache_id" json:"cache_id"`
	ExpirationDays       int64              `yaml:"expiration_days" json:"expiration_days"`
	MDMHost              string             `yaml:"mdm_host" json:"mdm_host"`
	MCConsumer           *MCConsumerConf    `yaml:"mc_consumer" json:"mc_consumer"`
	MDMAreas             map[string]bool    `yaml:"mdm_areas" json:"mdm_areas"`
	Executor             *ExecutorConf      `yaml:"executor" json:"executor"` // need to change to executor later
	MassiveConcurrency   int64              `yaml:"massive_concurrency" json:"massive_concurrency"`
}

type ExecutorConf struct {
	AppName        string `yaml:"app_name" json:"app_name"`
	AppVersion     string `yaml:"app_version" json:"app_version"`
	ChartVersion   string `yaml:"chart_version" json:"chart_version"`
	ExecutingJobID string `yaml:"excuting_job_id" json:"excuting_job_id"`
}

type RedisFailOverConf struct {
	// the prefix for sentinel service name
	Prefix       string `yaml:"prefix" json:"prefix"`
	Name         string `yaml:"name" json:"name"`
	SentinelPort string `yaml:"sentinelPort" json:"sentinelPort"`
	MasterName   string `yaml:"masterName" json:"masterName"`
}

type MCConsumerConf struct {
	NumOfJobs            int `yaml:"num_of_jobs" json:"num_of_jobs"`
	NumOfUniqueVehicles  int `yaml:"num_of_unique_vehicles" json:"num_of_unique_vehicles"`
	NumOfNormalLocations int `yaml:"num_of_normal_locations" json:"num_of_normal_locations"`
	NumOfFlexLocations   int `yaml:"num_of_flex_locations" json:"num_of_flex_locations"`
}

func (c *RedisFailOverConf) GetSentinelAddress() []string {
	logrus.Infof("prefix: %s, name: %s, sentinel port: %s", c.Prefix, c.Name, c.SentinelPort)
	addr := c.Prefix + c.Name + ":" + c.SentinelPort
	// repeat 6 times so that it's likely that we can connect to two or three of them
	return []string{addr, addr, addr, addr, addr, addr}

}

var Conf *Config

func Init() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	dat, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	Conf = &Config{}

	err = yaml.Unmarshal(dat, &Conf)
	if err != nil {
		panic(err)
	}

	if Conf.Namespace == "" {
		Conf.Namespace = "herald"
	}

	if Conf.ExpirationDays <= 0 {
		Conf.ExpirationDays = 7
	}

	if Conf.GatewayJWTToken == "" {
		panic("gateway_jwt_token missing")
	}
	if Conf.MatrixTimeoutSeconds <= 0 {
		Conf.MatrixTimeoutSeconds = 600
	}

	if Conf.MassiveConcurrency <= 0 {
		Conf.MassiveConcurrency = 1
	}

	if Conf.MCConsumer == nil {
		// Put very large numbers to avoid any MDM and on demand executor
		var mcConsumer = MCConsumerConf{
			NumOfJobs:            5000,
			NumOfUniqueVehicles:  1000,
			NumOfNormalLocations: 5000,
			NumOfFlexLocations:   5000,
		}
		Conf.MCConsumer = &mcConsumer
	}

	if Conf.Executor == nil {
		var executor = ExecutorConf{
			AppName:      "herald-executor",
			AppVersion:   "0.1.4-massive-0.44",
			ChartVersion: "0.1.2",
		}
		Conf.Executor = &executor
	}

	logrus.Info("successfully init config", Conf)
	validateConf()
}

// TODO: explore using annotation to do this automatically, refer to how nearby req dto is validated
func validateConf() {
	if Conf.RedisHost == "" && Conf.RedisRFSHost == nil {
		panic("empty RedisHost")
	}
	/*
		if Conf.GatewayHost == "" {
			panic("empty GatewayHost")
		}
	*/
}
