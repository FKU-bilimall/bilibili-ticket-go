package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func GetRandomMAC() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	mac := make([]byte, 6)
	r.Read(mac)
	mac[0] = (mac[0] | 2) & 0xfe
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}
