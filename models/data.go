package models

import (
	"errors"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type TicketData struct {
	ExpireTimestamp int64 `mapstructure:"expireTimestamp"`
	BuyerID         int64 `mapstructure:"buyerID"`
	ProjectID       int64 `mapstructure:"projectID"`
	SkuID           int64 `mapstructure:"skuID"`
	ScreenID        int64 `mapstructure:"screenID"`
}

type DataStorage struct {
	TicketData           *[]TicketData `mapstructure:"ticket"`
	ticketChangeCallback *func(storage *DataStorage, ticket TicketData)
	viper                *viper.Viper
	mutex                sync.Mutex
}

func NewDataStorage() (*DataStorage, error) {
	v := viper.New()
	v.SetConfigName("data")
	v.SetConfigType("json")
	v.AddConfigPath(".")
	v.SetDefault("ticket",
		&[]TicketData{},
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
	var configuration *DataStorage
	err = v.Unmarshal(&configuration)
	if err != nil {
		return nil, err
	}
	configuration.viper = v
	return configuration, nil
}

func (c *DataStorage) Save() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.viper.Set("ticket", &c.TicketData)
	err := c.viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}

func (c *DataStorage) AddTicket(data TicketData) {
	ts := c.GetTickets()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, ticket := range ts {
		if ticket.BuyerID == data.BuyerID && ticket.ProjectID == data.ProjectID && ticket.SkuID == data.SkuID && ticket.ScreenID == data.ScreenID {
			return
		}
	}
	*c.TicketData = append(*c.TicketData, data)
	if c.ticketChangeCallback != nil {
		(*c.ticketChangeCallback)(c, data)
	}
}

func (c *DataStorage) GetTickets() []TicketData {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	validTickets := make([]TicketData, 0)
	for _, ticket := range *c.TicketData {
		if time.Unix(ticket.ExpireTimestamp, 0).After(time.Now()) {
			validTickets = append(validTickets, ticket)
		}
	}
	c.TicketData = &validTickets
	return validTickets
}

func (c *DataStorage) RemoveTicket(index int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if index < 0 || index >= int64(len(*c.TicketData)) {
		return
	}
	*c.TicketData = append((*c.TicketData)[:index], (*c.TicketData)[index+1:]...)
}

func (c *DataStorage) SetTicketChangeNotifyFunc(f *func(storage *DataStorage, ticket TicketData)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.ticketChangeCallback = f
}
