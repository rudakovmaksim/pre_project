package configs

import (
	"project/internal/entities"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Configs struct {
	ActualTitles  []string
	URL           string
	Key           string
	ExchangeRates string
	ConnectString string
	Address       string
}

func LoadConfig() (*Configs, error) {
	cfg := &Configs{}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./") // C:/Users/Администратор/Desktop/project

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "error read config yaml file: %v", err)
	}

	if err := viper.MergeConfigMap(viper.GetStringMap("default")); err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "error merge configs from yaml file: %v", err)
	}

	cfg.ActualTitles = viper.GetStringSlice("service.actual_titles")

	cfg.URL = viper.GetString("client.url")
	cfg.Key = viper.GetString("client.key")
	cfg.ExchangeRates = viper.GetString("client.exchange_rates")

	cfg.ConnectString = viper.GetString("database.connect_string")

	cfg.Address = viper.GetString("server.address")

	return cfg, nil
}
