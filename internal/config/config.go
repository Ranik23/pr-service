package config

import (
	"fmt"
	
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Logging      LoggingConfig      `mapstructure:"logging"`
	Application  string             `mapstructure:"application"`
	PublicServer PublicServerConfig `mapstructure:"public_server"`
	Storage      StorageConfig      `mapstructure:"storage"`
}

func LoadConfig(configPath, envPath string) (*Config, error) {
	err := godotenv.Load(envPath)
	if err != nil {
		fmt.Printf("WARNING: error loading .env file from %s: %v\n", envPath, err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	err = viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading config file from %s: %v", configPath, err)
	}

	viper.AutomaticEnv()


	
	if err := viper.BindEnv("storage.postgres.hosts", "DB_HOST"); err != nil {
		return nil, fmt.Errorf("error binding env variable DB_HOST: %v", err)
	}
	if err := viper.BindEnv("storage.postgres.password", "DB_PASSWORD"); err != nil {
		return nil, fmt.Errorf("error binding env variable DB_PASSWORD: %v", err)
	}
	if err := viper.BindEnv("storage.redis.host", "REDIS_HOST"); err != nil {
		return nil, fmt.Errorf("error binding env variable REDIS_HOST: %v", err)
	}
	if err := viper.BindEnv("storage.redis.password", "REDIS_PASSWORD"); err != nil {
		return nil, fmt.Errorf("error binding env variable REDIS_PASSWORD: %v", err)
	}
	

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %v", err)
	}

	return &config, nil
}
