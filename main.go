package main

import (
	client "bilibili-ticket-go/bili"
	"bilibili-ticket-go/clock"
	"bilibili-ticket-go/global"
	"bilibili-ticket-go/models"
	"bilibili-ticket-go/models/bili/return"
	"bilibili-ticket-go/models/cookiejar"
	"bilibili-ticket-go/models/hooks"
	"bilibili-ticket-go/tui/keyboard"
	"bilibili-ticket-go/tui/primitives"
	tutils "bilibili-ticket-go/tui/utils"
	"bilibili-ticket-go/utils"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DeRuina/timberjack"
	"github.com/fatih/color"
	"github.com/gdamore/tcell/v2"
	"github.com/imroc/req/v3"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"

	_ "bilibili-ticket-go/captcha"
)

var (
	logger         = utils.GetLogger(global.GetLogger(), "main", nil)
	biliClient     *client.Client
	conf           *models.Configuration
	data           *models.DataStorage
	jar            *cookiejar.Jar
	app            *tview.Application
	loggerTextview *tview.TextView
	fileLogger     = &timberjack.Logger{
		Filename:         "logs/latest.log",
		MaxSize:          100, // megabytes
		MaxBackups:       30,  // backups
		MaxAge:           7,   // days
		Compress:         false,
		LocalTime:        true,
		BackupTimeFormat: "20060102-150405",
	}
	ticketOngoingInstance = ""
)

func init() {
	global.GetLogger().AddHook(hooks.NewLogFileRotateHook(fileLogger))
	if !utils.IsFileEmpty("logs/latest.log") {
		fileLogger.Rotate()
	}
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
	data, err = models.NewDataStorage()
	if err != nil {
		panic(err)
	}
	jar = cookiejar.New(&cookiejar.Options{
		PublicSuffixList: nil,
		DefaultCookies:   conf.Bilibili.Cookies,
	})
	biliClient = client.GetNewClient(jar, conf.Bilibili.BUVID, conf.Bilibili.RefreshToken, conf.Bilibili.Fingerprint, conf.Bilibili.InfocUUID)
	conf.Bilibili.BUVID = biliClient.GetBUVID()
	conf.Bilibili.Fingerprint = biliClient.GetFingerprint()
	conf.Bilibili.InfocUUID = biliClient.GetInfocUUID()
}

func main() {
	bc, err := clock.GetBilibiliClockOffset()
	if err != nil {
		logger.Warn("Failed to get Bilibili clock offset: ", err)
	} else {
		logger.Trace("The Offset Between You and Bilibili Server: ", bc)
	}
	ac, err := clock.GetNTPClockOffset("ntp.aliyun.com")
	if err != nil {
		logger.Warn("Failed to get NTP clock offset: ", err)
	} else {
		logger.Trace("The Offset Between You and Aliyun NTP Server: ", ac)
	}
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
	defer func() {
		data.Save()
	}()
	app = tview.NewApplication().EnableMouse(true).EnablePaste(true)
	mainPages := primitives.NewPages()
	functionPages := primitives.NewPages()
	functionPages.SetBorder(true).SetTitle("BILIBILI CLIENT")
	featureChoose := tview.NewFlex().SetDirection(tview.FlexRow)
	flex := tview.NewFlex().
		AddItem(featureChoose, 25, 1, false).
		AddItem(functionPages, 0, 4, false)
	k := keyboard.NewKeyboardCaptureInstance(app, flex)
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
				return
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
			var (
				tickets        []_return.TicketSkuScreenID
				selectedTicket _return.TicketSkuScreenID
				mutex          = sync.Mutex{} // Mutex to protect shared data
				projectID      string
				buyersID       []int64
				targetBuyer    int64
			)
			var (
				ticketList        *tview.DropDown
				buyerList         *tview.DropDown
				input             *tview.InputField
				addToQueueBtn     *tview.Button
				buyerSelectedFunc = func(text string, index int) {
					targetBuyer = buyersID[index]
					addToQueueBtn.SetDisabled(false)
				}
				ticketSelectFunc = func(text string, index int) {
					mutex.Lock()
					defer mutex.Unlock()
					selectedTicket = tickets[index]
					err, buyers := biliClient.GetBuyerNoSensitiveInfo()
					if err != nil {
						logger.Errorf("GetBuyerNoSensitiveInfo error: %v", err)
						tutils.PopupModal(fmt.Sprintf("Bilibili API Returned An Unexpected Value,\n%s", err), mainPages, map[string]func() bool{
							"OK": func() bool { return true },
						}, k)
						return
					}
					var buyerOptions []string
					var buyersIDA []int64
					for _, buyer := range buyers {
						buyerOptions = append(buyerOptions, fmt.Sprintf("%s-%s", buyer.Name, buyer.IdCard))
						buyersIDA = append(buyersID, buyer.Id)
					}
					buyersID = buyersIDA
					buyerList.SetOptions(buyerOptions, buyerSelectedFunc)
					buyerList.SetDisabled(false)
				}
				resetSelectionFunc = func() {
					if projectID == input.GetText() && projectID != "" {
						return
					}
					tickets = *new([]_return.TicketSkuScreenID)
					selectedTicket = *new(_return.TicketSkuScreenID)
					buyersID = []int64{}
					targetBuyer = -1
					ticketList.SetDisabled(true)
					buyerList.SetDisabled(true)
					addToQueueBtn.SetDisabled(true)
					ticketList.SetOptions([]string{"Nothing"}, nil)
					buyerList.SetOptions([]string{"Nothing"}, nil)
					ticketList.SetCurrentOption(0)
					buyerList.SetCurrentOption(0)
				}
				refreshTicketFunc = func() {
					mutex.Lock()
					defer mutex.Unlock()
					resetSelectionFunc()
					if input.GetText() == "" {
						return
					}
					var i []_return.TicketSkuScreenID
					if projectID == input.GetText() && projectID != "" {
						return
					}
					projectID = input.GetText()
					err, i = biliClient.GetTicketSkuIDsByProjectID(input.GetText())
					if err != nil {
						logger.Errorf("GetTicketSkuIDsByProjectID error: %v", err)
						tutils.PopupModal(fmt.Sprintf("Bilibili API Returned An Unexpected Value,\n%s", err), mainPages, map[string]func() bool{
							"OK": func() bool { return true },
						}, k)
						return
					}
					var options []string
					var validTickets []_return.TicketSkuScreenID
					for _, t := range i {
						if utils.IsTicketOnSale(t.Flags.Number) { //t.Flags.Number != 5 && t.Flags.Number != 3 && t.Flags.Number != 4 {
							validTickets = append(validTickets, t)
							options = append(options, fmt.Sprintf("%s-%s", t.Name, t.Desc))
						}
					}
					tickets = validTickets
					if len(validTickets) == 0 {
						logger.Errorf("No valid tickets found")
						tutils.PopupModal("No valid tickets found", mainPages, map[string]func() bool{
							"OK": func() bool { return true },
						}, k)
						return
					}
					ticketList.SetOptions(nil, nil)
					ticketList.SetOptions(options, ticketSelectFunc)
					ticketList.SetDisabled(false)
				}
				addToWaitingQueue = func() {
					mutex.Lock()
					defer mutex.Unlock()
					pid, err := strconv.ParseInt(projectID, 10, 64)
					if err != nil {
						return
					}
					data.AddTicket(models.TicketData{
						ExpireTimestamp: selectedTicket.SaleStat.End.Unix(),
						BuyerID:         targetBuyer,
						ProjectID:       pid,
						SkuID:           selectedTicket.SkuID,
						ScreenID:        selectedTicket.ScreenID,
					})
					tutils.PopupModal("Add to queue Successfully", mainPages, map[string]func() bool{
						"OK": func() bool { return true },
					}, k)
				}
			)
			root := tview.NewFlex().SetDirection(tview.FlexRow)
			buyerList = tview.NewDropDown().SetLabel("Select Buyer: ").SetOptions([]string{"Nothing"}, nil)
			ticketList = tview.NewDropDown().SetLabel("Select Ticket: ").SetOptions([]string{"Nothing"}, nil)
			ticketList.SetCurrentOption(0)
			buyerList.SetCurrentOption(0)
			ticketList.SetDisabledStyle(tcell.StyleDefault.Background(tcell.ColorLightBlue).Foreground(tview.Styles.ContrastSecondaryTextColor)).SetDisabled(true)
			buyerList.SetDisabledStyle(tcell.StyleDefault.Background(tcell.ColorLightBlue).Foreground(tview.Styles.ContrastSecondaryTextColor)).SetDisabled(true)
			input = tview.NewInputField().
				SetAcceptanceFunc(func(text string, ch rune) bool {
					_, err := strconv.Atoi(text)
					return err == nil
				}).
				SetLabel("Project ID: ").
				SetFieldWidth(20).
				SetPlaceholder("Enter Project ID").
				SetChangedFunc(func(text string) {
					mutex.Lock()
					defer mutex.Unlock()
					resetSelectionFunc()
				})
			input.SetDoneFunc(func(key tcell.Key) { refreshTicketFunc() })
			root.AddItem(tview.NewFlex().
				SetDirection(tview.FlexColumn).
				AddItem(input, 32, 1, false).
				AddItem(tview.NewBox(), 2, 0, false).
				AddItem(tview.NewButton("OK").SetSelectedFunc(refreshTicketFunc), 4, 0, false),
				1, 0, false)
			root.AddItem(tview.NewBox(), 1, 0, false)
			root.AddItem(ticketList, 1, 0, false)
			root.AddItem(tview.NewBox(), 1, 0, false)
			root.AddItem(buyerList, 1, 0, false)
			root.AddItem(tview.NewBox(), 1, 0, false)
			addToQueueBtn = tview.NewButton(" Add to Automatic Ticket Booking Queue ").SetSelectedFunc(addToWaitingQueue)
			addToQueueBtn.SetDisabled(true)
			root.AddItem(addToQueueBtn, 1, 0, false)
			functionPages.AddPage("ticket", root, true, false)
		}
		{
			root := tview.NewFlex()
			list := tview.NewList()
			list.AddItem("A", "", 0, nil)
			listFlex := tview.NewFlex().AddItem(list, 0, 1, true)
			listFlex.SetTitle("Pending List").SetBorder(true)
			root.AddItem(listFlex, 20, 0, false)
			detail := tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(tview.NewFlex().SetDirection(tview.FlexRow).SetBorder(true), 0, 1, false).
				AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
					AddItem(tview.NewBox(), 2, 0, false).
					AddItem(tview.NewButton("Cancel Task").SetDisabled(true), 0, 1, false).
					AddItem(tview.NewTextView().SetDynamicColors(true).SetMaxLines(200), 2, 0, false).
					AddItem(tview.NewButton("Force Start").SetDisabled(true), 0, 1, false).
					AddItem(tview.NewBox(), 2, 0, false), 1, 1, false)
			root.AddItem(detail, 0, 1, false)
			notify := func(storage *models.DataStorage, ticket models.TicketData) {
				logger.Debugf("ticket: %+v", ticket)
				//logger.Debugf("storage: %+v", storage)
			}
			data.SetTicketChangeNotifyFunc(&notify)
			functionPages.AddPage("status",
				root,
				true,
				false)
		}
		{
			root := tview.NewFlex()
			functionPages.AddPage("setting",
				root,
				true,
				false)
		}
	}
	{
		featureChoose.SetBorder(true).SetTitle("Features")
		{
			list := tview.NewList()
			list.AddItem("Bilibili Client", "Account Info/Login", 'l', func() {})
			list.AddItem("Logs", "Latest Logs", 'o', func() {})
			list.AddItem("Ticket", "Ticket Booking", 't', func() {})
			list.AddItem("Status", "Booking Status", 's', func() {})
			list.AddItem("Settings", "Configure", 'c', func() {})
			list.SetSelectedFunc(func(i int, mt string, _ string, _ rune) {
				functionPages.SetTitle(strings.ToUpper(mt))
				switch i {
				case 0:
					functionPages.SwitchToPage("client")
				case 1:
					functionPages.SwitchToPage("logs")
				case 2:
					functionPages.SwitchToPage("ticket")
				case 3:
					functionPages.SwitchToPage("status")
				case 4:
					functionPages.SwitchToPage("setting")
				}
			})
			featureChoose.AddItem(list, 0, 1, true)
		}
	}
	mainPages.AddPage("main", flex, true, true)
	app.SetInputCapture(k.InputCapture)
	app.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		if k.Selected() && (action == tview.MouseRightClick || action == tview.MouseMiddleClick || action == tview.MouseLeftClick) {
			k.Reset()
		}
		return event, action
	})
	go func() {
		logger.Info("It's Bilibili-Ticket-Go!!!!!")
		logger.Warnf("This is a %s Bilibili Client for ticket booking.", color.New(color.FgHiRed).Sprint("FREE"))
		logger.Info("Under the AGPLv3 License.")
		logger.Infof("Commit hash: %s", global.GitCommit)
		logger.Infof("Build timestamp: %s", global.BuildTime)
		err, r := biliClient.GetLoginStatus()
		if err != nil {
			logger.Errorf("Something went wrong when get logging status, %v", err)
		}
		if r.Login {
			err, b := biliClient.TryToRefreshNewBiliTicket()
			if err != nil {
				logger.Errorf("Something went wrong when refreshing bili-ticket, %v", err)
			} else if !b {
				logger.Info("No need to refresh bili-ticket, it is still valid.")
			} else {
				logger.Info("Bili-ticket refreshed successfully.")
			}
		}
	}()
	if err := app.SetRoot(mainPages, true).Run(); err != nil {
		logger.Fatal(err)
	}
}
