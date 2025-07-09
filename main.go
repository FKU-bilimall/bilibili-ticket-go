package main

import (
	client "bilibili-ticket-go/bili"
	"bilibili-ticket-go/models"
	"bilibili-ticket-go/models/cookiejar"
	"bilibili-ticket-go/utils"
	"fmt"
	"github.com/fatih/color"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"log"
	"net/url"
	"strings"
	"time"
)

var logger = utils.GetLogger("main", nil)
var biliClient *client.Client = nil
var conf *models.Configuration = nil
var jar *cookiejar.Jar = nil
var stack = (&models.Stack[selectItem]{}).New()
var selectedPrimitive selectItem
var app *tview.Application = nil
var container *tview.Flex = nil

type selectItem struct {
	FocusedNum int
	Obj        interface{}
}

func init() {
	var err error
	conf, err = models.NewConfiguration()
	if err != nil {
		panic(err)
	}
	jar = cookiejar.New(&cookiejar.Options{
		PublicSuffixList: nil,
		DefaultCookies:   conf.Bilibili.Cookies,
	})
	b, _ := url.ParseRequestURI("https://api.bilibili.com")
	a := jar.Cookies(b)
	if len(a) > 0 {

	}
	biliClient = client.GetNewClient(jar, conf.Bilibili.BUVID)
	conf.Bilibili.BUVID = biliClient.GetBUVID()
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	utils.RegisterLoggerFormater()
	defer func() {
		var ck = jar.AllPersistentEntries()
		if ck != nil {
			conf.Bilibili.Cookies = ck
		}
		conf.Save()
	}()
	app = tview.NewApplication().EnableMouse(true).EnablePaste(true)
	//app.SetInputCapture(keyCapture)
	container = tview.NewFlex()
	container.SetBorder(true).SetTitle("BILIBILI CLIENT")
	pages := tview.NewPages()
	{
		{
			t := tview.NewTextView()
			t.SetDynamicColors(true).
				SetScrollable(true).
				SetChangedFunc(func() {
					app.Draw()
				})
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
						for i := 0; i < 180; i++ {
							now := time.Now()
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
								etaWriter.Write([]byte(fmt.Sprintf("ETA: %ds left, ret-code: %d, msg: %s", 180-i, result.Code, result.Message)))
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
								} else {
									root.AddItem(btn, 3, 0, false)
								}
								return
							}
							time.Sleep(time.Duration(1)*time.Second - time.Since(now))
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
	}
	container.AddItem(pages, 0, 1, false)
	featureChoose := tview.NewFlex().SetDirection(tview.FlexRow)
	{
		featureChoose.SetBorder(true).SetTitle("Features")
		{
			list := tview.NewList()
			list.AddItem("Bilibili Client", "Account Info/Login", 'l', func() {})
			list.AddItem("Logs", "Latest Logs", 'o', func() {})
			list.SetSelectedFunc(func(i int, mt string, _ string, _ rune) {
				container.SetTitle(strings.ToUpper(mt))
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
		AddItem(container, 0, 4, false)
	selectedPrimitive = selectItem{FocusedNum: -1, Obj: nil}
	stack.Push(selectItem{FocusedNum: -1, Obj: container})
	go func() {
		logger.Info("It's Bilibili-Ticket-Go!!!!!")
		logger.Warn(fmt.Sprintf("This is a %s Bilibili Client for ticket booking.", color.New(color.FgHiRed).Sprint("FREE")))
		logger.Info("Under the AGPLv3 License.")
	}()
	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatal(err)
	}
}

func keyCapture(key *tcell.EventKey) *tcell.EventKey {
	if key.Key() == tcell.KeyTab {
		obj := stack.Top()
		switch obj := obj.Obj.(type) {
		case *tview.Flex:
			{
				s := stack.Top().FocusedNum + 1
				var count = 0
			CHECK:
				count++
				if s >= obj.GetItemCount() {
					if obj.GetItemCount() == 0 {
						return key
					}
					s = 0
				}
				item := obj.GetItem(s)
				if isAllowSkip(item) {
					if count == obj.GetItemCount() {
						return key
					}
					s++
					goto CHECK
				}
				selectedPrimitive = selectItem{
					FocusedNum: -1,
					Obj:        item,
				}
				top := stack.Top()
				stack.Pop()
				top.FocusedNum = s
				stack.Push(top)
			}
		}
	} else if key.Key() == tcell.KeyEnter {
		o, ok := selectedPrimitive.Obj.(tview.Primitive)
		if ok && selectedPrimitive != stack.Top() {
			app.SetFocus(o)
			switch obj := o.(type) {
			case *tview.Flex:
				if selectedPrimitive.FocusedNum == -1 {
					break
				}
				item := obj.GetItem(selectedPrimitive.FocusedNum)
				stack.Push(selectedPrimitive)
				selectedPrimitive = selectItem{
					FocusedNum: selectedPrimitive.FocusedNum,
					Obj:        item,
				}
				return key
			}
			stack.Push(selectedPrimitive)
			selectedPrimitive = selectItem{
				FocusedNum: -1,
				Obj:        nil,
			}
		}
	} else if key.Key() == tcell.KeyESC {
		if stack.Size() == 1 {
			app.SetFocus(container)
			return key
		}
		top := stack.Top()
		stack.Pop()
		item := stack.Top()
		o, ok := item.Obj.(tview.Primitive)
		if ok {
			selectedPrimitive = top
			app.SetFocus(o)
		}
	}
	return key
}

func setHighlight(oldObj, newObj interface{}) {
	switch old := oldObj.(type) {
	case *tview.List:
		old.SetBorderColor(tcell.ColorWhite)
	case *tview.Box:
		old.SetBorderColor(tcell.ColorWhite)
	case *tview.Flex:
		old.SetBorderColor(tcell.ColorWhite)
	case *tview.Grid:
		old.SetBorderColor(tcell.ColorWhite)
	case *tview.Button:
		old.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlue))
	}

	switch n := newObj.(type) {
	case *tview.List:
		n.SetBorderColor(tcell.NewHexColor(0x00cc00))
	case *tview.Box:
		n.SetBorderColor(tcell.NewHexColor(0x00cc00))
	case *tview.Flex:
		n.SetBorderColor(tcell.NewHexColor(0x00cc00))
	case *tview.Grid:
		n.SetBorderColor(tcell.NewHexColor(0x00cc00))
	case *tview.Button:
		n.SetStyle(tcell.StyleDefault.Background(tcell.NewHexColor(0xB0C4DE)))
	}
}

func isAllowSkip(pri tview.Primitive) bool {
	switch pri.(type) {
	case *tview.TextView:
		return true
	default:
		return false
	}
}
