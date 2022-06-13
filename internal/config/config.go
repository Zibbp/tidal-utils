package config

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

type Conf struct {
	Debug   bool
	Manual  bool
	Spotify struct {
		ClientID     string
		ClientSecret string
		AccessToken  string
		RefreshToken string
		Expiry       time.Time
		TokenType    string
		RedirectURI  string
	}
	Tidal struct {
		UserID       string
		AccessToken  string
		RefreshToken string
	}
}

func NewConfig() {
	configLocation := "/data"
	configName := "config"
	configType := "json"
	configPath := fmt.Sprintf("%s/%s.%s", configLocation, configName, configType)

	viper.AddConfigPath("/data")
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	viper.SetDefault("debug", false)
	viper.SetDefault("manual", false)
	viper.SetDefault("spotify.client_id", "")
	viper.SetDefault("spotify.client_secret", "")
	viper.SetDefault("spotify.access_token", "")
	viper.SetDefault("spotify.refresh_token", "")
	viper.SetDefault("spotify.expiry", "")
	viper.SetDefault("spotify.token_type", "")
	viper.SetDefault("spotify.redirect_uri", "http://localhost:28542/callback")
	viper.SetDefault("tidal.user_id", "")
	viper.SetDefault("tidal.access_token", "")
	viper.SetDefault("tidal.refresh_token", "")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Info("Config file not found, creating...")
		err := viper.SafeWriteConfigAs(configPath)
		if err != nil {
			log.Panicf("Error creating config file: %w", err)
		}
	} else {
		err := viper.ReadInConfig()
		if err != nil {
			log.Errorf("Error reading config file: %w", err)
		}
	}
}
