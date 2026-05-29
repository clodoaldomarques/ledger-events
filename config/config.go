package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"
)

type Config struct {
	AppPort            int
	MySqlDBUser        string
	MySqlDBPass        string
	MySqlDBHost        string
	MySqlDBPort        string
	MysqlDBName        string
	AwsAddress         string
	AwsRegion          string
	AwsAccessKeyID     string
	AwsSecretAccessKey string
	EventTopic         string
	LedgerConfigApiUrl string
}

type Option func(*Config)

var (
	singleton sync.Once
	instance  *Config
)

func New(options ...Option) *Config {
	singleton.Do(func() {
		instance = &Config{
			AppPort:            GetInt("APP_PORT", 8080),
			MySqlDBUser:        GetString("MYSQL_USER", "admin"),
			MySqlDBPass:        GetString("MYSQL_PASSWORD", "l3dg3r"),
			MySqlDBHost:        GetString("MYSQL_HOST", "192.168.49.2"),
			MySqlDBPort:        GetString("MYSQL_PORT", "30001"),
			MysqlDBName:        GetString("MYSQL_DATABASE", "ledger"),
			AwsAddress:         GetString("AWS_ADDRESS", "http://192.168.49.2:30002"),
			AwsRegion:          GetString("AWS_REGION", "us-east-1"),
			AwsAccessKeyID:     GetString("AWS_ACCESS_KEY_ID", "test"),
			AwsSecretAccessKey: GetString("AWS_SECRET_ACCESS_KEY", "test"),
			EventTopic:         GetString("EVENT_SNS_TOPIC", "events-sns-topic"),
			LedgerConfigApiUrl: GetString("LEDGER_CONFIG_API_URL", "http://192.168.49.2:31000"),
		}
	})

	for _, optFunc := range options {
		optFunc(instance)
	}

	return instance
}

func WithAppPort(appPort int) Option {
	return func(c *Config) {
		c.AppPort = appPort
	}
}

func WithAwsAddress(awsAddress string) Option {
	return func(c *Config) {
		c.AwsAddress = awsAddress
	}
}
func WithAwsRegion(awsRegion string) Option {
	return func(c *Config) {
		c.AwsRegion = awsRegion
	}
}

func WithLedgerConfigApiUrl(ledgerConfigApiUrl string) Option {
	return func(c *Config) {
		c.LedgerConfigApiUrl = ledgerConfigApiUrl
	}
}

func (c Config) Region() string {
	return c.AwsRegion
}

func (c Config) Address() string {
	return c.AwsAddress
}
func (c Config) AccessKeyID() string {
	return c.AwsAccessKeyID
}
func (c Config) SecretAccessKey() string {
	return c.AwsSecretAccessKey
}

func (c Config) TopicARN() string {
	return c.EventTopic
}

func GetString(env string, def string) string {
	if e := os.Getenv(env); e != "" {
		return e
	}
	return def
}

func GetInt(env string, def int) int {
	i, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		return def
	}
	return i
}

func (c Config) GetMySQLConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		c.MySqlDBUser,
		c.MySqlDBPass,
		c.MySqlDBHost,
		c.MySqlDBPort,
		c.MysqlDBName,
	)
}
