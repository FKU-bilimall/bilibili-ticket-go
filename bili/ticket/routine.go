package ticket

import (
	client "bilibili-ticket-go/bili"
	"bilibili-ticket-go/bili/token"
	"bilibili-ticket-go/global"
	"bilibili-ticket-go/models"
	"bilibili-ticket-go/models/bili/api"
	r "bilibili-ticket-go/models/bili/return"
	"bilibili-ticket-go/models/enums"
	"bilibili-ticket-go/models/hooks"
	"bilibili-ticket-go/utils"
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Routine struct {
	mutex     sync.Mutex
	client    *client.Client
	buyer     r.TicketBuyer
	ticket    models.TicketEntry
	ctx       context.Context
	isRunning bool
	cancel    context.CancelFunc
	logger    *logrus.Logger
}

func NewTicketRoutine(client *client.Client, buyer r.TicketBuyer, ticket models.TicketEntry, h []logrus.Hook) *Routine {
	ctx, cancel := context.WithCancel(context.Background())
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	level, err := strconv.Atoi(global.LoggerLevel)
	if err != nil {
		level = 4
	}
	logger.SetLevel(logrus.Level(level))
	for _, hook := range h {
		logger.AddHook(hook)
	}
	tr := &Routine{
		client:    client,
		isRunning: false,
		buyer:     buyer,
		ticket:    ticket,
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
	}
	logger.AddHook(hooks.NewRoutineHandlerHook(func(i int, fields logrus.Fields) {
		if i == enums.Success || i == enums.Failed || i == enums.Error {
			tr.isRunning = false
		}
	}))
	utils.RegisterLoggerFormater(logger)
	return tr
}

func (tr *Routine) Start() {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	if tr.isRunning {
		return
	}

	tr.isRunning = true
	go run(tr.client, tr.buyer, tr.ticket, 500*time.Millisecond, tr.ctx, tr.logger)
}

func (tr *Routine) IsRunning() bool {
	return tr.isRunning
}

func (tr *Routine) Stop() {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	if !tr.isRunning {
		return
	}

	tr.cancel()
	tr.isRunning = false
}

func run(client *client.Client, buyer r.TicketBuyer, ticketData models.TicketEntry, interval time.Duration, ctx context.Context, mainLog *logrus.Logger) {
	pidString := strconv.FormatInt(ticketData.ProjectID, 10)
	logger := utils.GetLogger(mainLog, fmt.Sprintf("%s", ticketData.Hash()[:11]), nil)
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
			if buyer.BuyerType == enums.Ordinary {
				buyerInterface = map[string]string{
					"tel":  buyer.Tel,
					"name": buyer.Name,
				}
			} else if buyer.BuyerType == enums.ForceRealName {
				err, confirm := client.GetConfirmInformation(tk, strconv.FormatInt(ticketData.ProjectID, 10))
				if err != nil {
					logger.WithField("status", enums.Error).WithError(err).Errorf("GetConfirmInformation err: %v", err)
					goto SLEEP
				}
				for _, b := range confirm.BuyerList.List {
					if b.Id == buyer.ID {
						buyerInterface = b
					}
				}
				if buyerInterface == nil {
					logger.WithField("status", enums.Error).WithError(err).Errorf("GetConfirmInformation buyer not found in project %d", ticketData.ProjectID)
					return
				}
			}
			err, code, msg, to = client.SubmitOrder(tokenGen, whenGenPtoken, tk, pidString, *ticket, buyerInterface, buyer.BuyerType)
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
