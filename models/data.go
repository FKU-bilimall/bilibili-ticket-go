package models

import (
	_return "bilibili-ticket-go/models/bili/return"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type TicketEntry struct {
	Expire      int64
	Start       int64
	ProjectID   int64
	ProjectName string
	SkuID       int64
	SkuName     string
	ScreenID    int64
	ScreenName  string
	Buyer       _return.TicketBuyer
}

func (t TicketEntry) String() string {
	return t.ProjectName + " - " + t.SkuName + " - " + t.ScreenName + " - " + t.Buyer.String() + " (Expire: " + time.Unix(t.Expire, 0).Format("2006-01-02 15:04:05") + ";Start: " + time.Unix(t.Start, 0).Format("2006-01-02 15:04:05") + ")"
}

func (t TicketEntry) Hash() string {
	str := fmt.Sprintf(
		"Buyer:BuyerType:%d,ID:%d,Name:%s,Tel:%s|Expire:%d|Start:%d|ProjectID:%d|ScreenID:%d|SkuID:%d",
		t.Buyer.BuyerType, t.Buyer.ID, t.Buyer.Name, t.Buyer.Tel, t.Expire, t.Start, t.ProjectID, t.ScreenID, t.SkuID)
	hash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash[:])
}

func (t TicketEntry) Valid() bool {
	return t.Expire > time.Now().Unix() && t.ProjectID > 0 && t.SkuID > 0 && t.ScreenID > 0 && t.Buyer.Valid()
}

type DataStorage struct {
	TicketData           []TicketEntry `mapstructure:"ticket"`
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
		[]TicketEntry{},
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
	c.TicketData = append(c.TicketData, data)
	if c.ticketChangeCallback != nil {
		go func() {
			(*c.ticketChangeCallback)(c, data)
		}()
	}
}

func (c *DataStorage) GetTickets() []TicketEntry {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	validTickets := make([]TicketEntry, 0)
	for _, ticket := range c.TicketData {
		if time.Unix(ticket.Expire, 0).After(time.Now()) {
			validTickets = append(validTickets, ticket)
		}
	}
	c.TicketData = validTickets
	return validTickets
}

func (c *DataStorage) RemoveTicket(index int64) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if index < 0 || index >= int64(len(c.TicketData)) {
		return false
	}
	old := c.TicketData[index]
	c.TicketData = append((c.TicketData)[:index], (c.TicketData)[index+1:]...)
	if c.ticketChangeCallback != nil {
		go func() {
			(*c.ticketChangeCallback)(c, old)
		}()
	}
	return true
}

func (c *DataStorage) RemoveTicketByHash(hash string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for i, ticket := range c.TicketData {
		if ticket.Hash() == hash {
			old := ticket
			c.TicketData = append((c.TicketData)[:i], (c.TicketData)[i+1:]...)
			if c.ticketChangeCallback != nil {
				go func() {
					(*c.ticketChangeCallback)(c, old)
				}()
			}
			return true
		}
	}
	return false
}
func (c *DataStorage) SetTicketChangeNotifyFunc(f *func(storage *DataStorage, ticket TicketEntry)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.ticketChangeCallback = f
}
