package config

import (
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	HTTPport               string `envconfig:"API_PORT"`
	DBhost                 string `envconfig:"DB_HOST"`
	DBname                 string `envconfig:"POSTGRES_DB"`
	DBusername             string `envconfig:"POSTGRES_USER"`
	DBpassword             string `envconfig:"POSTGRES_PASSWORD"`
	AccessTokenKey         string `envconfig:"ACCESS_KEY"`
	AccessTokenExpMinutes  int    `envconfig:"ACCESS_EXP_MINUTES"`
	RefreshTokenExpMinutes int    `envconfig:"REFRESH_EXP_MINUTES"`
	RefreshTokenLen        int    `envconfig:"REFRESH_TOKEN_LEN"`
	MathcingKeyLen         int    `envconfig:"MATCHING_KEY_LEN"`
}

func GetConfig() (*config, error) {
	cfg := &config{}
	err := envconfig.Process("", cfg)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}
