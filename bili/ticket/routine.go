package ticket

import (
	client "bilibili-ticket-go/bili"
	"bilibili-ticket-go/bili/token"
	"bilibili-ticket-go/global"
	"bilibili-ticket-go/models"
	"bilibili-ticket-go/models/bili/api"
	r "bilibili-ticket-go/models/bili/return"
	"bilibili-ticket-go/models/enums"
	"bilibili-ticket-go/utils"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type TicketRoutine struct {
	mutex      sync.Mutex
	client     *client.Client
	buyer      r.BuyerInformation //if is "-1", doesn't use any buyer
	ticket     models.TicketEntry
	ctx        context.Context
	isRunning  bool
	cancel     context.CancelFunc
	loggerHook logrus.Hook
}

func NewTicketRoutine(client *client.Client, buyer r.BuyerInformation, ticket models.TicketEntry) *TicketRoutine {
	ctx, cancel := context.WithCancel(context.Background())
	return &TicketRoutine{
		client:    client,
		isRunning: false,
		buyer:     buyer,
		ticket:    ticket,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (tr *TicketRoutine) Start() {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	if tr.isRunning {
		return
	}

	tr.isRunning = true
	go run(tr.client, tr.buyer, tr.ticket, 500*time.Millisecond, tr.ctx)
}

func (tr *TicketRoutine) Stop() {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	if !tr.isRunning {
		return
	}

	tr.cancel()
	tr.isRunning = false
}

func run(client *client.Client, buyerID r.BuyerInformation, ticketData models.TicketEntry, interval time.Duration, ctx context.Context) {
	var bid string
	if buyerID.ContactInfo != nil {
		bid = buyerID.ContactInfo.Tel
	} else {
		bid = strconv.FormatInt(buyerID.ForceRealNameBuyer.Id, 10)
	}
	pidString := strconv.FormatInt(ticketData.ProjectID, 10)
	logger := utils.GetLogger(global.GetLogger(), fmt.Sprintf("%#X-%#X-%#X-%s", ticketData.ProjectID, ticketData.SkuID, ticketData.ScreenID, bid), nil)
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
	var buyer *api.BuyerStruct
	var count uint16 = 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var (
				err  error
				code int
				msg  string
				to   api.TicketOrderStruct
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
			if buyer == nil && buyerID.ForceRealNameBuyer != nil {
				err, confirm := client.GetConfirmInformation(tk, strconv.FormatInt(ticketData.ProjectID, 10))
				if err != nil {
					logger.WithField("status", enums.Error).WithError(err).Errorf("GetConfirmInformation err: %v", err)
					goto SLEEP
				}
				for _, b := range confirm.BuyerList.List {
					if b.Id == buyerID.ForceRealNameBuyer.Id {
						buyer = &b
					}
				}
				if buyer == nil {
					logger.WithField("status", enums.Error).WithError(err).Errorf("GetConfirmInformation buyer not found in project %d", ticketData.ProjectID)
					return
				}
			}
			err, code, msg, to = client.SubmitOrder(tokenGen, whenGenPtoken, tk, pidString, *ticket, r.BuyerInformation{
				ForceRealNameBuyer: buyer,
				ContactInfo:        buyerID.ContactInfo,
			})
			if err != nil {
				logger.WithField("status", enums.Error).WithError(err).Errorf("SubmitOrder err: %v", err)
				goto SLEEP
			}
			if (code == 0 || code == 100048 || code == 100079) && (to.OrderId != 0 && to.OrderCreateTime != 0) {
				// 肘击成功
				logger.WithFields(logrus.Fields{
					"status": enums.Success,
					"bili": logrus.Fields{
						"code":    code,
						"message": msg,
					},
				}).Infof("SubmitOrder success, orderID: %d", to.OrderId)
				return
			} else if code == 100034 {
				// 价格不对捏
				ticket.Price = to.PayMoney
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
