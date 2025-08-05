package models

import (
	"bilibili-ticket-go/models/cookiejar"
	"errors"

	"github.com/spf13/viper"
)

type Bilibili struct {
	RefreshToken string
	Cookies      []cookiejar.CookieEntries
	BUVID        string
	Fingerprint  string
}
type Configuration struct {
	Bilibili *Bilibili `mapstructure:"bilibili"`
	viper    *viper.Viper
}

func NewConfiguration() (*Configuration, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("json")
	v.AddConfigPath(".")
	v.SetDefault("bilibili",
		&Bilibili{
			Cookies: make([]cookiejar.CookieEntries, 0),
		},
	)
	err := v.SafeWriteConfig()
	if err != nil {
		var configFileAlreadyExistsError viper.ConfigFileAlreadyExistsError
		if !errors.As(err, &configFileAlreadyExistsError) {
			return nil, err
		}
	}
	err = v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	var configuration *Configuration
	err = v.Unmarshal(&configuration)
	if err != nil {
		return nil, err
	}
	configuration.viper = v
	return configuration, nil
}
func (c *Configuration) Save() error {
	c.viper.Set("bilibili", &c.Bilibili)
	err := c.viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}
