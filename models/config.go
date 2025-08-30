package models

import (
	"bilibili-ticket-go/bili"
	"bilibili-ticket-go/models/cookiejar"
	"errors"

	"github.com/spf13/viper"
)

type Bilibili struct {
	RefreshToken string
	Cookies      []cookiejar.CookieEntries
	BUVID        string
	InfocUUID    string
	Fingerprint  bili.Fingerprint
}

type TicketSetting struct {
	AutoStartBuying bool   `mapstructure:"autoStartBuying"`
	NtpServer       string `mapstructure:"ntpServer"`
}

type Configuration struct {
	Bilibili *Bilibili      `mapstructure:"bilibili"`
	Ticket   *TicketSetting `mapstructure:"ticket"`
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
	v.SetDefault("ticket",
		&TicketSetting{
			AutoStartBuying: false,
			NtpServer:       "ntp.aliyun.com",
		})
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
