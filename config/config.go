package config

import (
	"fmt"
	"sync"

	"github.com/clodoaldomarques/core-sdk/pkg/env"
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
			AppPort:            env.GetInt("APP_PORT", 8080),
			MySqlDBUser:        env.GetString("MYSQL_USER", ""),
			MySqlDBPass:        env.GetString("MYSQL_PASSWORD", ""),
			MySqlDBHost:        env.GetString("MYSQL_HOST", ""),
			MySqlDBPort:        env.GetString("MYSQL_PORT", ""),
			MysqlDBName:        env.GetString("MYSQL_DATABASE", ""),
			AwsAddress:         env.GetString("AWS_ADDRESS", ""),
			AwsRegion:          env.GetString("AWS_REGION", ""),
			AwsAccessKeyID:     env.GetString("AWS_ACCESS_KEY_ID", ""),
			AwsSecretAccessKey: env.GetString("AWS_SECRET_ACCESS_KEY", ""),
			EventTopic:         env.GetString("EVENTS_SNS_TOPIC", ""),
			LedgerConfigApiUrl: env.GetString("LEDGER_CONFIG_API_URL", ""),
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

func (c Config) GetMySQLConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		c.MySqlDBUser,
		c.MySqlDBPass,
		c.MySqlDBHost,
		c.MySqlDBPort,
		c.MysqlDBName,
	)
}
