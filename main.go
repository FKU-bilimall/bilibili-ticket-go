package main

import (
	client "bilibili-ticket-go/bili"
	"bilibili-ticket-go/utils"
	"fmt"
	cookiejar "github.com/juju/persistent-cookiejar"
	"log"
	"time"
)

func main() {
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: "cookies.json",
	})
	if err != nil {
		log.Fatalf("failed to create persistent cookiejar: %s\n", err.Error())
	}
	defer jar.Save()
	c := client.GetNewClient(jar, "")
	err, qr := c.GetQRCodeUrlAndKey()
	if err != nil {
		return
	}
	asciiBits, _ := utils.GetQRCode(qr.URL, false)
	for _, bit := range asciiBits {
		fmt.Println(bit)
	}
	for i := 0; i < 120; i++ {
		err, data := c.GetQRLoginState(qr.QRCodeKey)
		if err != nil {
			return
		}
		fmt.Println(data)
		time.Sleep(1 * time.Second)
		if data.Code == 0 {
			return
		}
	}
}
