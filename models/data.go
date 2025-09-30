package models

import (
	"bilibili-ticket-go/models/enums"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type TicketEntry struct {
	Expire    int64       `mapstructure:"expire"`
	ProjectID int64       `mapstructure:"projectID"`
	Project   string      `mapstructure:"projectName"`
	SkuID     int64       `mapstructure:"skuID"`
	Sku       string      `mapstructure:"skuName"`
	ScreenID  int64       `mapstructure:"screenID"`
	Screen    string      `mapstructure:"screenName"`
	Buyer     TicketBuyer `mapstructure:"buyer"`
}

type TicketBuyer struct {
	BuyerType enums.BuyerType `mapstructure:"type"`
	ID        int64           `mapstructure:"ID,omitempty"`
	Tel       string          `mapstructure:"tel,omitempty"`
	Name      string          `mapstructure:"name"`
}

func (buyer TicketBuyer) Compare(a TicketBuyer) bool {
	if buyer.BuyerType != a.BuyerType {
		return false
	}
	if buyer.BuyerType == enums.Ordinary {
		return buyer.Tel == a.Tel && buyer.Name == a.Name
	} else {
		return buyer.ID == a.ID
	}
}

func (buyer TicketBuyer) String() string {
	if buyer.BuyerType == enums.Ordinary {
		return buyer.Name + " (" + buyer.Tel + ")"
	} else {
		return buyer.Name + " (ID: " + strconv.FormatInt(buyer.ID, 10) + ")"
	}
}

type DataStorage struct {
	TicketData           *[]TicketEntry `mapstructure:"ticket"`
	ticketChangeCallback *func(storage *DataStorage, ticket TicketEntry)
	viper                *viper.Viper
	mutex                sync.Mutex
}

func NewDataStorage() (*DataStorage, error) {
	v := viper.New()
	v.SetConfigName("data")
	v.SetConfigType("json")
	v.AddConfigPath(".")
	v.SetDefault("ticket",
		&[]TicketEntry{},
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

func (c *DataStorage) AddTicket(data TicketEntry) {
	ts := c.GetTickets()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, ticket := range ts {
		if ticket.Buyer.Compare(data.Buyer) && ticket.ProjectID == data.ProjectID && ticket.SkuID == data.SkuID && ticket.ScreenID == data.ScreenID {
			return
		}
	}
	*c.TicketData = append(*c.TicketData, data)
	if c.ticketChangeCallback != nil {
		(*c.ticketChangeCallback)(c, data)
	}
}

func (c *DataStorage) GetTickets() []TicketEntry {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	validTickets := make([]TicketEntry, 0)
	for _, ticket := range *c.TicketData {
		if time.Unix(ticket.Expire, 0).After(time.Now()) {
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

func (c *DataStorage) SetTicketChangeNotifyFunc(f *func(storage *DataStorage, ticket TicketEntry)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.ticketChangeCallback = f
}
