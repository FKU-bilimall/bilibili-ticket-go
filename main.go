package main

import (
	client "bilibili-ticket-go/bili"
	"bilibili-ticket-go/models"
	"bilibili-ticket-go/models/cookiejar"
	"bilibili-ticket-go/utils"
	"fmt"
	"github.com/fatih/color"
	"github.com/imroc/req/v3"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"log"
	"strings"
	"time"
)

var logger = utils.GetLogger("main", nil)
var biliClient *client.Client = nil
var conf *models.Configuration = nil
var jar *cookiejar.Jar = nil
var app *tview.Application = nil
var pageContainer *tview.Flex = nil

func init() {
	req.SetDefaultClient(req.DefaultClient().SetLogger(utils.GetLogger("network", nil)).EnableDebugLog())
	var err error
	conf, err = models.NewConfiguration()
	if err != nil {
		panic(err)
	}
	jar = cookiejar.New(&cookiejar.Options{
		PublicSuffixList: nil,
		DefaultCookies:   conf.Bilibili.Cookies,
	})
	biliClient = client.GetNewClient(jar, conf.Bilibili.BUVID, conf.Bilibili.RefreshToken)
	conf.Bilibili.BUVID = biliClient.GetBUVID()
}

func main() {
	logrus.SetLevel(logrus.TraceLevel)
	utils.RegisterLoggerFormater()
	defer func() {
		var ck = jar.AllPersistentEntries()
		if ck != nil {
			conf.Bilibili.Cookies = ck
		}
		conf.Save()
	}()
	app = tview.NewApplication().EnableMouse(true).EnablePaste(true)
	pageContainer = tview.NewFlex()
	pageContainer.SetBorder(true).SetTitle("BILIBILI CLIENT")
	pages := tview.NewPages()
	{
		{
			t := tview.NewTextView()
			t.SetDynamicColors(true).
				SetScrollable(true).
				SetMaxLines(2000).
				SetChangedFunc(func() {
					app.Draw()
				})
			t.ScrollToEnd()
			logrus.SetOutput(tview.ANSIWriter(t))
			pages.AddPage("logs",
				tview.NewFlex().SetDirection(tview.FlexRow).AddItem(t, 0, 1, false),
				true,
				false)
		}
		{
			root := tview.NewFlex().SetDirection(tview.FlexRow)
			t := tview.NewTextView()
			t.SetChangedFunc(func() {
				app.Draw()
			})
			root.AddItem(t, 2, 0, false)
			err, stat := biliClient.GetLoginStatus()
			if err != nil {
				logrus.Errorf("GetLoginStatus error: %v", err)
			}
			if stat.Login {
				t.Write([]byte(fmt.Sprintf("Welcome %s, Your UID is %d", stat.Name, stat.UID)))
				err, _, s := biliClient.CheckAndUpdateCookie()
				if err != nil {
					logger.Errorf("CheckAndUpdateCookie error: %v", err)
				} else if s != "" {
					conf.Bilibili.RefreshToken = s
				}
			} else {
				t.Write([]byte("You are not logged in. Please login first."))
				qrv := tview.NewTextView().SetChangedFunc(func() { app.Draw() })
				eta := tview.NewTextView().SetChangedFunc(func() { app.Draw() })
				etaWriter := tview.ANSIWriter(eta)
				btn := tview.NewButton("Get QR Code")
				btn.SetBorder(true)
				btn.SetSelectedFunc(func() {
					root.RemoveItem(btn)
					root.RemoveItem(eta)
					qrv.Clear()
					err, d := biliClient.GetQRCodeUrlAndKey()
					var expire = time.Now().Add(179 * time.Second)
					if err != nil {
						logger.Errorf("GetQRCodeUrlAndKey error: %v", err)
					}
					qr, _ := utils.GetQRCode(d.URL, false)
					for i, s := range qr {
						if i == len(qr)-1 {
							qrv.Write([]byte(s))
						} else {
							qrv.Write([]byte(s + "\n"))
						}
					}
					root.AddItem(qrv, 0, 1, false)
					root.AddItem(eta, 1, 0, false)
					go func() {
						timer := time.NewTimer(1 * time.Second)
					FOR:
						for {
							select {
							case <-timer.C:
								var now = time.Now()
								err, result := biliClient.GetQRLoginState(d.QRCodeKey)
								if err != nil {
									logger.Errorf("GetQRLoginState error: %v", err)
								}
								if result.Code == 86038 {
									root.RemoveItem(eta)
									root.RemoveItem(qrv)
									eta.Clear()
									etaWriter.Write([]byte(fmt.Sprintf("Qrcode is expired, please get a new one.")))
									root.AddItem(btn, 3, 0, false)
									root.AddItem(eta, 1, 0, false)
									return
								}
								if result.Code != 0 {
									eta.Clear()
									etaWriter.Write([]byte(fmt.Sprintf("ETA: %.0fs left, ret-code: %d, msg: %s", (expire.Sub(now)).Seconds(), result.Code, result.Message)))
								}
								if result.Code == 0 {
									eta.Clear()
									root.RemoveItem(eta)
									root.RemoveItem(qrv)
									err, stat := biliClient.GetLoginStatus()
									conf.Bilibili.RefreshToken = result.RefreshToken
									conf.Bilibili.LatestRefreshTimestamp = result.Timestamp
									if err != nil {
										logrus.Errorf("GetLoginStatus error: %v", err)
									}
									if stat.Login {
										t.Clear()
										t.Write([]byte(fmt.Sprintf("Welcome %s, Your UID is %d", stat.Name, stat.UID)))
									} else {
										root.AddItem(btn, 3, 0, false)
									}
									break FOR
								}
								offest := time.Now().Sub(now)
								if offest.Seconds() > 1 {
									offest = 1 * time.Second
								}
								timer.Reset(1*time.Second - offest)
							}
						}
						root.RemoveItem(eta)
						root.RemoveItem(qrv)
						eta.Clear()
						etaWriter.Write([]byte(fmt.Sprintf("Qrcode is expired, please get a new one.")))
						root.AddItem(btn, 3, 0, false)
						root.AddItem(eta, 1, 0, false)
						return
					}()

				})
				root.AddItem(btn, 3, 0, false)
			}
			pages.AddPage("client", root, true, true)
		}
		{
		}
	}
	pageContainer.AddItem(pages, 0, 1, false)
	featureChoose := tview.NewFlex().SetDirection(tview.FlexRow)
	{
		featureChoose.SetBorder(true).SetTitle("Features")
		{
			list := tview.NewList()
			list.AddItem("Bilibili Client", "Account Info/Login", 'l', func() {})
			list.AddItem("Logs", "Latest Logs", 'o', func() {})
			list.AddItem("Ticket", "Ticket Booking", 't', func() {})
			list.SetSelectedFunc(func(i int, mt string, _ string, _ rune) {
				pageContainer.SetTitle(strings.ToUpper(mt))
				switch i {
				case 0:
					pages.SwitchToPage("client")
				case 1:
					pages.SwitchToPage("logs")
				}
			})
			featureChoose.AddItem(list, 0, 1, true)
		}
	}
	flex := tview.NewFlex().
		AddItem(featureChoose, 25, 1, false).
		AddItem(pageContainer, 0, 4, false)
	keyboard := NewKeyboardCaptureInstance(app, flex)
	app.SetInputCapture(keyboard.InputCapture)
	go func() {
		logger.Info("It's Bilibili-Ticket-Go!!!!!")
		logger.Warn(fmt.Sprintf("This is a %s Bilibili Client for ticket booking.", color.New(color.FgHiRed).Sprint("FREE")))
		logger.Info("Under the AGPLv3 License.")
		biliClient.RefreshNewBiliTicket()
	}()
	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatal(err)
	}
}
