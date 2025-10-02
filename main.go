package main

import (
	client "bilibili-ticket-go/bili"
	"bilibili-ticket-go/bili/ticket"
	"bilibili-ticket-go/clock"
	"bilibili-ticket-go/global"
	"bilibili-ticket-go/models"
	"bilibili-ticket-go/models/bili/api"
	"bilibili-ticket-go/models/bili/return"
	"bilibili-ticket-go/models/cookiejar"
	"bilibili-ticket-go/models/enums"
	"bilibili-ticket-go/models/hooks"
	"bilibili-ticket-go/notify"
	"bilibili-ticket-go/scheduler"
	"bilibili-ticket-go/tui/keyboard"
	"bilibili-ticket-go/tui/primitives"
	tutils "bilibili-ticket-go/tui/utils"
	"bilibili-ticket-go/utils"
	"fmt"
	"os"
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

type ticketRoutineInformation struct {
	routine  *ticket.Routine
	logCache *hooks.LoggerCache
}

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
		Compression:      "none",
		LocalTime:        true,
		BackupTimeFormat: "20060102-150405",
	}
	ticketRoutineInfo               = make(map[string]*ticketRoutineInformation)
	successTicketTask               = make(map[string]bool)
	schedulerManager                = scheduler.NewDynamicScheduler()
	notifyManager     notify.Notify = nil
)

func init() {
	global.GetLogger().AddHook(hooks.NewLogFileRotateHook(fileLogger))
	if st, err := os.Stat("logs/latest.log"); !utils.IsFileEmpty("logs/latest.log") && err != nil && st.Size() >= int64(fileLogger.MaxSize)*1000*1000 {
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
	switch enums.ConvertNotificationType(conf.Ticket.Notification.Type) {
	case enums.None:
		notifyManager = nil
	case enums.Gotify:
		notifyManager = notify.NewGotify(conf.Ticket.Notification.Token, conf.Ticket.Notification.Endpoint)
	}
}

func main() {
	/*	bc, err := clock.GetBilibiliClockOffset()
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
		}*/
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
		for s, b := range successTicketTask {
			if b {
				data.RemoveTicketByHash(s)
			}
		}
		data.Save()
	}()
	defer func() {
		if p := recover(); p != nil {
			if app != nil {
				app.Stop()
			}
			panic(p)
		}
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
				projID         string
				buyers         []api.BuyerNoSensitiveStruct
				buyer          api.BuyerNoSensitiveStruct
				buyerType      = enums.ForceRealName
				projName       string
			)
			var (
				ticketList        = primitives.NewDropDown()
				buyerList         = primitives.NewDropDown()
				buyerTelInput     = primitives.NewInputField()
				buyerNameInput    = primitives.NewInputField()
				buyerFlex         = tview.NewFlex().SetDirection(tview.FlexRow)
				input             *tview.InputField
				addToQueueBtn     *tview.Button
				buyerSelectedFunc = func(text string, index int) {
					buyer = buyers[index]
					addToQueueBtn.SetDisabled(false)
				}
				buyerCheckFunc = func() {
					if buyerNameInput.GetText() != "" && buyerTelInput.GetText() != "" {
						addToQueueBtn.SetDisabled(false)
					} else {
						addToQueueBtn.SetDisabled(true)
					}
				}
				buyerFlexResetFunc = func() {
					buyerFlex.Clear()
					buyerFlex.AddItem(primitives.NewLabel("Buyer: Waiting for Specific Project"), 1, 0, false)
				}
				ticketSelectFunc = func(text string, index int) {
					mutex.Lock()
					defer mutex.Unlock()
					selectedTicket = tickets[index]
					if buyerType == enums.ForceRealName {
						err, res := biliClient.GetBuyerNoSensitiveInfo()
						if err != nil {
							logger.Errorf("GetBuyerNoSensitiveInfo error: %v", err)
							tutils.PopupModal(fmt.Sprintf("Bilibili API Returned An Unexpected Value,\n%s", err), mainPages, map[string]func() bool{
								"OK": func() bool { return true },
							}, k)
							return
						}
						var buyerOptions []string
						for _, buyer := range res {
							buyerOptions = append(buyerOptions, fmt.Sprintf("%s-%s", buyer.Name, buyer.IdCard))
						}
						buyers = res
						buyerList.SetOptions(buyerOptions, buyerSelectedFunc)
						buyerList.SetDisabled(false)
					} else if buyerType == enums.Ordinary {
						buyerNameInput.SetDisabled(false)
						buyerTelInput.SetDisabled(false)
					}
				}
				resetSelectionFunc = func() {
					if projID == input.GetText() && projID != "" {
						return
					}
					tickets = nil
					selectedTicket = _return.TicketSkuScreenID{}
					buyers = []api.BuyerNoSensitiveStruct{}
					buyer = api.BuyerNoSensitiveStruct{}
					ticketList.SetDisabled(true)
					buyerList.SetDisabled(true)
					addToQueueBtn.SetDisabled(true)
					buyerNameInput.SetDisabled(true)
					buyerTelInput.SetDisabled(true)
					ticketList.SetOptions([]string{"Nothing"}, nil)
					buyerList.SetOptions([]string{"Nothing"}, nil)
					ticketList.SetCurrentOption(0)
					buyerList.SetCurrentOption(0)
					buyerFlexResetFunc()
				}
				refreshTicketFunc = func() {
					mutex.Lock()
					defer mutex.Unlock()
					resetSelectionFunc()
					if input.GetText() == "" {
						return
					}
					err, info := biliClient.GetProjectInformation(input.GetText())
					projName = info.ProjectName
					buyerFlex.Clear()
					if info.IsNeedContact {
						buyerType = enums.Ordinary
						buyerFlex.AddItem(buyerNameInput, 1, 0, false)
						buyerFlex.AddItem(tview.NewBox(), 1, 0, false)
						buyerFlex.AddItem(buyerTelInput, 1, 0, false)
					} else {
						buyerType = enums.ForceRealName
						buyerFlex.AddItem(tview.NewBox(), 1, 0, false)
						buyerFlex.AddItem(buyerList, 1, 0, false)
					}
					var i []_return.TicketSkuScreenID
					if projID == input.GetText() && projID != "" {
						return
					}
					projID = input.GetText()
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
						buyerFlexResetFunc()
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
					pid, err := strconv.ParseInt(projID, 10, 64)
					if err != nil {
						return
					}
					entry := models.TicketEntry{
						ProjectID:   pid,
						ProjectName: projName,
						Expire:      selectedTicket.SaleStat.End.Unix(),
						Start:       selectedTicket.SaleStat.Start.Unix(),
						SkuID:       selectedTicket.SkuID,
						SkuName:     selectedTicket.Desc,
						ScreenID:    selectedTicket.ScreenID,
						ScreenName:  selectedTicket.Name,
					}
					if buyerType == enums.Ordinary {
						if buyerNameInput.GetText() == "" || buyerTelInput.GetText() == "" {
							tutils.PopupModal("Please enter complete contact information", mainPages, map[string]func() bool{
								"OK": func() bool { return true },
							}, k)
							return
						}
						entry.Buyer = _return.TicketBuyer{
							BuyerType: enums.Ordinary,
							Tel:       buyerTelInput.GetText(),
							Name:      buyerNameInput.GetText(),
						}
					} else {
						if buyer.Id == 0 {
							tutils.PopupModal("Please select a buyer", mainPages, map[string]func() bool{
								"OK": func() bool { return true },
							}, k)
							return
						}
						entry.Buyer = _return.TicketBuyer{
							BuyerType: enums.ForceRealName,
							ID:        buyer.Id,
							Name:      buyer.Name,
						}
					}
					data.AddTicket(entry)
					tutils.PopupModal("Add to queue Successfully", mainPages, map[string]func() bool{
						"OK": func() bool { return true },
					}, k)
				}
			)
			{
				buyerTelInput.SetLabel("Tel: ").SetFieldWidth(20).SetPlaceholder("Enter Tel")
				buyerNameInput.SetLabel("Name: ").SetFieldWidth(20).SetPlaceholder("Enter Name")
				ticketList.SetLabel("Select Ticket: ").SetOptions([]string{"Nothing"}, nil).SetCurrentOption(0)
				buyerList.SetLabel("Select Buyer: ").SetOptions([]string{"Nothing"}, nil).SetCurrentOption(0)
				ticketList.SetDisabledStyle(tcell.StyleDefault.Background(tcell.ColorLightBlue).Foreground(tview.Styles.ContrastSecondaryTextColor)).SetDisabled(true)
				buyerList.SetDisabledStyle(tcell.StyleDefault.Background(tcell.ColorLightBlue).Foreground(tview.Styles.ContrastSecondaryTextColor)).SetDisabled(true)
				buyerNameInput.SetDisabledStyle(tcell.StyleDefault.Background(tcell.ColorLightBlue).Foreground(tview.Styles.ContrastSecondaryTextColor)).SetDisabled(true)
				buyerTelInput.SetDisabledStyle(tcell.StyleDefault.Background(tcell.ColorLightBlue).Foreground(tview.Styles.ContrastSecondaryTextColor)).SetDisabled(true)
				buyerFlexResetFunc()
				buyerNameInput.SetChangedFunc(func(text string) { buyerCheckFunc() })
				buyerTelInput.SetChangedFunc(func(text string) { buyerCheckFunc() })
			}
			root := tview.NewFlex().SetDirection(tview.FlexRow)
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
			root.AddItem(buyerFlex, 3, 0, false)
			root.AddItem(tview.NewBox(), 1, 0, false)
			addToQueueBtn = tview.NewButton(" Add to Automatic Ticket Booking Queue ").SetSelectedFunc(addToWaitingQueue)
			addToQueueBtn.SetDisabled(true)
			root.AddItem(addToQueueBtn, 1, 0, false)
			functionPages.AddPage("ticket", root, true, false)
		}
		{
			var (
				current = -1
				hash    []string
			)
			root := primitives.NewPages()
			root.SetBorder(true).SetTitle("TICKET LIST")
			list := tview.NewList()
			logs := tview.NewTextView().SetMaxLines(200).SetDynamicColors(true).SetChangedFunc(func() {
				if app != nil {
					app.Draw()
				}
			})
			logFlex := tview.NewFlex().AddItem(logs, 0, 1, true).SetDirection(tview.FlexRow)
			logFlex.SetBorder(true)
			detail := tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(logFlex, 0, 1, false).
				AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
					AddItem(tview.NewBox(), 2, 0, false).
					AddItem(tview.NewButton("Exit").SetSelectedFunc(func() {
						root.SwitchToPage("list")
						root.SetTitle("TICKET LIST")
					}), 0, 1, false).
					AddItem(tview.NewBox(), 2, 0, false).
					AddItem(tview.NewButton("Cancel Task").SetSelectedFunc(func() {
						tutils.PopupModal("Are you sure delete this task?", mainPages, map[string]func() bool{
							"Yes": func() bool {
								root.SwitchToPage("list")
								root.SetTitle("TICKET LIST")
								if current != -1 {
									if ticketRoutineInfo[hash[current]].routine.IsRunning() {
										ticketRoutineInfo[hash[current]].routine.Stop()
									}
									data.RemoveTicket(int64(current))
								}
								return true
							},
							"No": func() bool { return true },
						}, k)
					}), 0, 1, false).
					AddItem(tview.NewTextView().SetDynamicColors(true).SetMaxLines(200), 2, 0, false).
					AddItem(tview.NewButton("Force Start").SetSelectedFunc(func() {
						successTicketTask[hash[current]] = false
						if ticketRoutineInfo[hash[current]].routine.IsRunning() {
							ticketRoutineInfo[hash[current]].routine.Stop()
						}
						ticketRoutineInfo[hash[current]].routine.Start()
					}), 0, 1, false).
					AddItem(tview.NewBox(), 2, 0, false), 1, 0, false)
			ANSI := tview.ANSIWriter(logs)
			notify := func(storage *models.DataStorage, t models.TicketEntry) {
				list.Clear()
				hash = []string{}
				logger.Debugf("t: %+v", t)
				for i, t := range storage.GetTickets() {
					h := t.Hash()
					if _, exists := ticketRoutineInfo[h]; exists {
						continue
					}
					cache := hooks.NewLoggerCache(200, nil)
					handler := hooks.NewRoutineHandlerHook(func(i int, fields logrus.Fields) {
						if i == enums.Success || i == enums.Failed || i == enums.Error {
							schedulerManager.RemoveTask(h)
							successTicketTask[h] = true
						}
					})
					loghooks := []logrus.Hook{cache, handler}
					err, routine := ticket.NewTicketRoutine(biliClient, t, loghooks, notifyManager)
					if err != nil {
						logger.Errorf("Failed to create ticket routine[hash:%s]: %v", h[:11], err)
						continue
					}
					ticketRoutineInfo[h] = &ticketRoutineInformation{
						routine:  routine,
						logCache: cache,
					}
					schedulerManager.AddTask(h, time.Unix(t.Expire, 0), func() {
						if !ticketRoutineInfo[hash[current]].routine.IsRunning() {
							ticketRoutineInfo[hash[current]].routine.Start()
						}
					})
					hash = append(hash, h)
					list.AddItem(fmt.Sprintf("%d.%s(%s)", i+1, t.ProjectName, t.SkuName), fmt.Sprintf(" [%s]{%s}(%s)", t.ScreenName, t.Buyer.Name, h[0:9]), 0, nil)
				}
				logger.Debugf("storage: %+v", storage)
			}
			data.SetTicketChangeNotifyFunc(&notify)
			notify(data, models.TicketEntry{})
			functionPages.AddPage("status",
				root,
				true,
				false)
			list.SetSelectedFunc(func(i int, _ string, _ string, _ rune) {
				if current != -1 {
					ticketRoutineInfo[hash[current]].logCache.SetOutput(nil)
				}
				current = i
				logs.Clear()
				for _, s := range ticketRoutineInfo[hash[current]].logCache.GetEntries() {
					ANSI.Write([]byte(s))
				}
				logs.ScrollToEnd()
				ticketRoutineInfo[hash[current]].logCache.SetOutput(ANSI)
				root.SwitchToPage("detail")
				root.SetTitle("DETAIL")
			})
			list.SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
				if list.GetItemCount() == 0 {
					// Draw a horizontal line across the middle of the box.
					centerY := y + height/2
					for cx := x + 1; cx < x+width-1; cx++ {
						screen.SetContent(cx, centerY, tview.BoxDrawingsLightHorizontal, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
					}

					// Write som text along the horizontal line.
					tview.Print(screen, " Empty List ", x+1, centerY, width-2, tview.AlignCenter, tcell.ColorYellow)
				}
				// Space for other content.
				return x, y, width, height
			})
			root.AddPage("list", list, true, true)
			root.AddPage("detail", detail, true, false)
		}
		{
			root := tview.NewFlex()
			form := tview.NewForm()
			form.AddButton("Save", nil).
				AddButton("Reset", func() {
					app.Stop()
				})
			root.AddItem(form, 0, 1, true)
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
	//offest sync
	go func() {
		for {
			ac, err := clock.GetNTPClockOffset("ntp.aliyun.com")
			if err != nil {
				continue
			}
			schedulerManager.SetGlobalOffset(ac)
			time.Sleep(60 * time.Second) //setting
		}
	}()
	if err := app.SetRoot(mainPages, true).Run(); err != nil {
		logger.Fatal(err)
	}
}
