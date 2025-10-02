package ticket

import (
	client "bilibili-ticket-go/bili"
	"bilibili-ticket-go/bili/token"
	"bilibili-ticket-go/global"
	"bilibili-ticket-go/models"
	"bilibili-ticket-go/models/bili/api"
	r "bilibili-ticket-go/models/bili/return"
	"bilibili-ticket-go/models/enums"
	"bilibili-ticket-go/models/errors"
	"bilibili-ticket-go/models/hooks"
	"bilibili-ticket-go/notify"
	"bilibili-ticket-go/utils"
	"context"
	errors2 "errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/DeRuina/timberjack"
	"github.com/sirupsen/logrus"
)

type Routine struct {
	mutex     sync.RWMutex
	client    *client.Client
	buyer     r.TicketBuyer
	ticket    models.TicketEntry
	ctx       context.Context
	isRunning bool
	cancel    context.CancelFunc
	logger    *logrus.Entry
}

func NewTicketRoutine(client *client.Client, ticket models.TicketEntry, h []logrus.Hook, notify notify.Notify) (error, *Routine) {
	if !ticket.Valid() {
		return errors.NewRoutineCreateError("ticket data is invalid"), nil
	}
	if client == nil {
		return errors.NewRoutineCreateError("bili-client is nil"), nil
	}
	err, info := client.GetLoginStatus()
	if err != nil {
		return errors2.Join(errors.NewRoutineCreateError("get login status error"), err), nil
	}
	hash := ticket.Hash()
	ctx, cancel := context.WithCancel(context.Background())
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	level, err := strconv.Atoi(global.LoggerLevel)
	if err != nil {
		level = 4
	}
	logger.SetLevel(logrus.Level(level))
	fileLogger := &timberjack.Logger{
		Filename:    "logs/tickets/" + hash[0:11] + ".log",
		MaxSize:     1e5, // megabytes
		MaxBackups:  0,   // backups
		MaxAge:      7,   // days
		Compression: "none",
		LocalTime:   true,
	}
	logger.AddHook(hooks.NewLogFileRotateHook(fileLogger))
	for _, hook := range h {
		logger.AddHook(hook)
	}
	entry := utils.GetLogger(logger, fmt.Sprintf("%s", hash[0:11]), nil)
	tr := &Routine{
		client:    client,
		isRunning: false,
		ticket:    ticket,
		ctx:       ctx,
		cancel:    cancel,
		logger:    entry,
	}
	logger.AddHook(hooks.NewRoutineHandlerHook(func(i int, fields logrus.Fields) {
		if i == enums.Success || i == enums.Failed || i == enums.Error {
			entry.Info("Ticket Routine stopped")
			tr.setIsRunning(false)
		}
		if i == enums.Success && notify != nil {
			notify.Notify(fmt.Sprintf("抢票成功！\n项目：%s\n场次：%s\n票种：%s\n购票人：%s\n购票用户：%s(%d)", ticket.ProjectName, ticket.ScreenName, ticket.SkuName, ticket.Buyer.String(), info.Name, info.UID))
		}
	}))
	utils.RegisterLoggerFormater(logger)
	entry.Info("Ticket Routine created")
	return nil, tr
}

func (tr *Routine) Start() {
	tr.logger.Info("Ticket Routine started")
	if tr.IsRunning() {
		return
	}

	tr.setIsRunning(true)
	go run(tr.client, tr.ticket, 500*time.Millisecond, tr.ctx, tr.logger)
}

func (tr *Routine) setIsRunning(val bool) {
	//tr.mutex.Lock()
	//defer tr.mutex.Unlock()
	tr.isRunning = val
}

func (tr *Routine) IsRunning() bool {
	//tr.mutex.RLock()
	//defer tr.mutex.RUnlock()
	return tr.isRunning
}

func (tr *Routine) Stop() {
	tr.logger.Info("Ticket Routine stopped")
	if !tr.IsRunning() {
		return
	}

	tr.cancel()
	tr.setIsRunning(false)
}

func run(client *client.Client, ticketData models.TicketEntry, interval time.Duration, ctx context.Context, logger *logrus.Entry) {
	pidString := strconv.FormatInt(ticketData.ProjectID, 10)
	err, info := client.GetProjectInformation(pidString)
	if err != nil {
		logger.WithField("status", enums.Error).WithError(err).Error("GetProjectInformation err")
		return
	}
	var tokenGen token.Generator
	if info.IsHotProject {
		tokenGen = token.NewCTokenGenerator()
	} else {
		tokenGen = token.NewNormalTokenGenerator()
	}
	err, tickets := client.GetTicketSkuIDsByProjectID(pidString)
	if err != nil {
		logger.WithField("status", enums.Error).WithError(err).Errorf("GetTicketSkuIDsByProjectID err: %v", err)
		return
	}
	var ticket *r.TicketSkuScreenID
	for _, t := range tickets {
		if t.SkuID == ticketData.SkuID && t.ScreenID == ticketData.ScreenID {
			ticket = &t
			break
		}
	}
	if ticket == nil {
		logger.WithField("status", enums.Error).WithError(err).Errorf("Ticket with skuID %d not found in project %d", ticketData.SkuID, ticketData.ProjectID)
		return
	}
	whenGenPtoken := time.Now()
	err, tk := client.GetRequestTokenAndPToken(tokenGen, strconv.FormatInt(ticketData.ProjectID, 10), *ticket)
	if err != nil {
		logger.WithField("status", enums.Error).WithError(err).Errorf("GetRequestTokenAndPToken err: %v", err)
		return
	}
	var count uint16 = 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var (
				err            error
				code           int
				msg            string
				to             api.TicketOrderStruct
				buyerInterface interface{}
			)
			if count >= 61 {
				// 该换个新token去骗叔叔了
				whenGenPtoken = time.Now()
				err, tk = client.GetRequestTokenAndPToken(tokenGen, strconv.FormatInt(ticketData.ProjectID, 10), *ticket)
				if err != nil {
					logger.WithField("status", enums.Error).WithError(err).Errorf("GetRequestTokenAndPToken err: %v", err)
				}
				goto SLEEP
			}
			if ticketData.Buyer.BuyerType == enums.Ordinary {
				buyerInterface = map[string]string{
					"tel":  ticketData.Buyer.Tel,
					"name": ticketData.Buyer.Name,
				}
			} else if ticketData.Buyer.BuyerType == enums.ForceRealName {
				err, confirm := client.GetConfirmInformation(tk, strconv.FormatInt(ticketData.ProjectID, 10))
				if err != nil {
					logger.WithField("status", enums.Error).WithError(err).Errorf("GetConfirmInformation err: %v", err)
					goto SLEEP
				}
				for _, b := range confirm.BuyerList.List {
					if b.Id == ticketData.Buyer.ID {
						buyerInterface = b
					}
				}
				if buyerInterface == nil {
					logger.WithField("status", enums.Error).WithError(err).Errorf("GetConfirmInformation buyer not found in project %d", ticketData.ProjectID)
					return
				}
			}
			err, code, msg, to = client.SubmitOrder(tokenGen, whenGenPtoken, tk, pidString, *ticket, buyerInterface, ticketData.Buyer.BuyerType)
			if err != nil {
				logger.WithField("status", enums.Error).WithError(err).Errorf("SubmitOrder err: %v", err)
				goto SLEEP
			}
			if (code == 0 || code == 100048 || code == 100079) && to.OrderId != 0 {
				// 肘击成功
				logger.WithFields(logrus.Fields{
					"status": enums.Success,
					"bili": logrus.Fields{
						"code":    code,
						"message": msg,
						"order":   to.OrderId,
					},
				}).Infof("SubmitOrder success, orderID: %d", to.OrderId)
				return
			} else if code == 100034 {
				// 价格不对捏
				ticket.Price = to.PayMoney
			} else if code == 100017 {
				// 不可售
				logger.WithFields(logrus.Fields{
					"status": enums.Failed,
					"bili": logrus.Fields{
						"code":    code,
						"message": msg,
					},
				}).Warnf("%s (%d)", msg, code)
				return
			}
			logger.WithFields(logrus.Fields{
				"status": enums.Pending,
				"bili": logrus.Fields{
					"code":    code,
					"message": msg,
				},
			}).Infof("%s (%d)", msg, code)
		SLEEP:
			count++
			time.Sleep(interval)
		}
	}
}
