package main

import (
	client "bilibili-ticket-go/bili"
	"bilibili-ticket-go/bili/clock"
	"bilibili-ticket-go/global"
	"bilibili-ticket-go/keyboard"
	"bilibili-ticket-go/models"
	"bilibili-ticket-go/models/cookiejar"
	"bilibili-ticket-go/models/hooks"
	"bilibili-ticket-go/tui"
	"bilibili-ticket-go/utils"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DeRuina/timberjack"
	"github.com/fatih/color"
	"github.com/gdamore/tcell/v2"
	"github.com/imroc/req/v3"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

var logger = utils.GetLogger(global.GetLogger(), "main", nil)
var biliClient *client.Client = nil
var conf *models.Configuration = nil
var jar *cookiejar.Jar = nil
var app *tview.Application = nil
var loggerTextview *tview.TextView = nil
var fileLogger = &timberjack.Logger{
	Filename:   "logs/latest.log", // Choose an appropriate path
	MaxSize:    100,               // megabytes
	MaxBackups: 10,                // backups
	MaxAge:     7,                 // days
	Compress:   false,             // default: false
	LocalTime:  true,              // default: false (use UTC)
	//RotationInterval: time.Hour * 24,    // Rotate daily if no other rotation met
	BackupTimeFormat: "20060102-150405", // Rotated files will have format <logfilename>-2006-01-02-15-04-05-<rotationCriterion>-timberjack.log
}

func init() {
	global.GetLogger().AddHook(hooks.NewLogFileRotateHook(fileLogger))
	fileLogger.Rotate()
	loggerTextview = tview.NewTextView()
	loggerTextview.SetDynamicColors(true).
		SetScrollable(true).
		SetMaxLines(2000).
		SetChangedFunc(func() {
			if app != nil {
				app.Draw()
			}
		})
	global.GetLogger().SetOutput(tview.ANSIWriter(loggerTextview))
	req.SetDefaultClient(req.DefaultClient().SetLogger(utils.GetLogger(global.GetLogger(), "network", nil)).EnableDebugLog())
	var err error
	conf, err = models.NewConfiguration()
	if err != nil {
		panic(err)
	}
	jar = cookiejar.New(&cookiejar.Options{
		PublicSuffixList: nil,
		DefaultCookies:   conf.Bilibili.Cookies,
	})
	biliClient = client.GetNewClient(jar, conf.Bilibili.BUVID, conf.Bilibili.RefreshToken, conf.Bilibili.Fingerprint)
	conf.Bilibili.BUVID = biliClient.GetBUVID()
	conf.Bilibili.Fingerprint = biliClient.GetFingerprint()
}

func main() {
	bc, _ := clock.GetBilibiliClockOffset()
	ac, _ := clock.GetAliyunClockOffset()
	logger.Trace("The Offest Between You and Bilibili Server: ", bc)
	logger.Trace("The Offest Between You and Aliyun NTP Server: ", ac)
	defer fileLogger.Close()
	defer func() {
		var ck = jar.AllPersistentEntries()
		if ck != nil {
			conf.Bilibili.Cookies = ck
		}
		t := biliClient.GetRefreshToken()
		if t != "" {
			conf.Bilibili.RefreshToken = t
		}
		conf.Save()
	}()
	app = tview.NewApplication().EnableMouse(true).EnablePaste(true)
	mainPages := tui.NewPages()
	functionPages := tui.NewPages()
	functionPages.SetBorder(true).SetTitle("BILIBILI CLIENT")
	{
		{
			loggerTextview.ScrollToEnd()
			functionPages.AddPage("logs",
				tview.NewFlex().SetDirection(tview.FlexRow).AddItem(loggerTextview, 0, 1, false),
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
				err, f := biliClient.CheckAndUpdateCookie()
				if f {
					logger.Trace("Refresh cookie successfully.")
				}
				if err != nil {
					logger.Errorf("CheckAndUpdateCookie error: %v", err)
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
						b := false
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
									if err != nil {
										logrus.Errorf("GetLoginStatus error: %v", err)
									}
									if stat.Login {
										t.Clear()
										t.Write([]byte(fmt.Sprintf("Welcome %s, Your UID is %d", stat.Name, stat.UID)))
										b = true
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
						if b {
							return
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
			functionPages.AddPage("client", root, true, true)
		}
		{
			root := tview.NewFlex().SetDirection(tview.FlexRow)

			input := tview.NewInputField().
				SetAcceptanceFunc(func(text string, ch rune) bool {
					_, err := strconv.Atoi(text)
					return err == nil
				}).
				SetLabel("Project ID: ").
				SetFieldWidth(20).
				SetPlaceholder("Enter Project ID")
			list := tview.NewList().ShowSecondaryText(false)
			tickets := tview.NewFlex().SetDirection(tview.FlexColumn).AddItem(list, 0, 1, false)
			var refreshFunc = func() {
				list.Clear()
				logger.Info("Project ID: ", input.GetText())
				err, i, _ := biliClient.GetTicketSkuIDsByProjectID(input.GetText())
				if err != nil {
					logger.Errorf("GetTicketSkuIDsByProjectID error: %v", err)
					return
				}
				for _, t := range i {
					if t.Flags.Number != 5 && t.Flags.Number != 3 {
						list.AddItem(fmt.Sprintf("%s-%s", t.Name, t.Desc), "", 0, nil)
					}
				}
			}
			input.SetFinishedFunc(func(key tcell.Key) { refreshFunc() })
			root.AddItem(tview.NewFlex().
				SetDirection(tview.FlexColumn).
				AddItem(input, 32, 1, false).
				AddItem(tview.NewBox(), 2, 0, false).
				AddItem(tview.NewButton("OK").SetSelectedFunc(func() {
					modal := tview.NewModal().SetText("确定要退出吗？").AddButtons([]string{"确定", "取消"})
					modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						if buttonIndex == 0 {
							app.Stop()
						} else {
							mainPages.RemovePage("quitModalPage")
						}
					})
					mainPages.AddPage("quitModalPage", modal, true, true)
				}), 4, 0, false),
				1, 0, false)
			root.AddItem(tickets, 0, 0, false)
			functionPages.AddPage("ticket", root, true, false)
		}
	}
	featureChoose := tview.NewFlex().SetDirection(tview.FlexRow)
	{
		featureChoose.SetBorder(true).SetTitle("Features")
		{
			list := tview.NewList()
			list.AddItem("Bilibili Client", "Account Info/Login", 'l', func() {})
			list.AddItem("Logs", "Latest Logs", 'o', func() {})
			list.AddItem("Ticket", "Ticket Booking", 't', func() {})
			list.SetSelectedFunc(func(i int, mt string, _ string, _ rune) {
				functionPages.SetTitle(strings.ToUpper(mt))
				switch i {
				case 0:
					functionPages.SwitchToPage("client")
				case 1:
					functionPages.SwitchToPage("logs")
				case 2:
					functionPages.SwitchToPage("ticket")
				}
			})
			featureChoose.AddItem(list, 0, 1, true)
		}
	}
	flex := tview.NewFlex().
		AddItem(featureChoose, 25, 1, false).
		AddItem(functionPages, 0, 4, false)
	mainPages.AddPage("main", flex, true, true)
	k := keyboard.NewKeyboardCaptureInstance(app, mainPages)
	app.SetInputCapture(k.InputCapture)
	app.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		if k.Selected() && (action == tview.MouseRightClick || action == tview.MouseMiddleClick || action == tview.MouseLeftClick) {
			k.Reset()
		}
		return event, action
	})
	go func() {
		logger.Info("It's Bilibili-Ticket-Go!!!!!")
		logger.Warn(fmt.Sprintf("This is a %s Bilibili Client for ticket booking.", color.New(color.FgHiRed).Sprint("FREE")))
		logger.Info("Under the AGPLv3 License.")
		err, r := biliClient.GetLoginStatus()
		if err != nil {
			logger.Errorf("Something went wrong when get logging status, %v", err)
		}
		if r.Login {
			err, b := biliClient.RefreshNewBiliTicket()
			if err != nil {
				logger.Errorf("Something went wrong when refreshing bili-ticket, %v", err)
			} else if !b {
				logger.Info("No need to refresh bili-ticket, it is still valid.")
			} else {
				logger.Info("Bili-ticket refreshed successfully.")
			}
		}
	}()
	if err := app.SetRoot(flex, true).Run(); err != nil {
		logger.Fatal(err)
	}
}
